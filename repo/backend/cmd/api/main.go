package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fitcommerce/internal/bootstrap"
	"fitcommerce/internal/jobs"
	"fitcommerce/internal/platform"
)

func main() {
	cfgVal := platform.LoadConfig()
	cfg := &cfgVal

	logger := platform.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	logger.Info("starting fitcommerce api server",
		"port", cfg.ServerPort,
		"log_level", cfg.LogLevel,
	)

	tlsCertFile := cfg.TLSCertFile
	tlsKeyFile := cfg.TLSKeyFile
	tlsCleanup := func() {}
	if !cfg.AllowInsecureHTTP {
		var err error
		tlsCertFile, tlsKeyFile, tlsCleanup, err = platform.EnsureServerTLSFiles(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			logger.Error("TLS is required but certificate setup failed", "error", err)
			os.Exit(1)
		}
		defer tlsCleanup()
		if cfg.TLSCertFile == "" && cfg.TLSKeyFile == "" {
			logger.Warn("no TLS certificate configured; generated an ephemeral self-signed certificate for local secure transport")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := bootstrap.NewApp(ctx, *cfg, logger, nil)
	if err != nil {
		logger.Error("failed to bootstrap application", "error", err)
		os.Exit(1)
	}
	defer app.Close()
	logger.Info("database connection established")

	go jobs.NewAutoCloseJob(app.Services.Order, logger).Run(ctx)
	go jobs.NewCutoffEvalJob(app.Services.Campaign, logger).Run(ctx)
	go jobs.NewBackupJob(app.Services.Backup, logger).Run(ctx)
	go jobs.NewVarianceDeadlineJob(app.Services.Variance, logger).Run(ctx)
	go jobs.NewRetentionCleanupJob(app.Services.Retention, logger).Run(ctx)
	if cfg.BiometricModuleEnabled {
		go jobs.NewBiometricKeyRotationJob(app.Services.Biometric, logger, cfg.BiometricKeyRotationDays).Run(ctx)
	}

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	useTLS := !cfg.AllowInsecureHTTP || (cfg.TLSCertFile != "" && cfg.TLSKeyFile != "")

	go func() {
		if useTLS {
			logger.Info("starting server with tls", "addr", addr)
			if err := app.Echo.StartTLS(addr, tlsCertFile, tlsKeyFile); err != nil && err != http.ErrServerClosed {
				logger.Error("server error", "error", err)
				os.Exit(1)
			}
		} else {
			logger.Info("starting server", "addr", addr)
			if err := app.Echo.Start(addr); err != nil && err != http.ErrServerClosed {
				logger.Error("server error", "error", err)
				os.Exit(1)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("received shutdown signal", "signal", sig.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := app.Echo.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited gracefully")
}
