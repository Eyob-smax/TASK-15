package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"

	"fitcommerce/internal/application"
	internalhttp "fitcommerce/internal/http"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/store/postgres"
)

// AppServices exposes the runtime services needed by jobs and integration tests.
type AppServices struct {
	Campaign  application.CampaignService
	Order     application.OrderService
	Backup    application.BackupService
	Retention application.RetentionService
	Variance  application.VarianceService
	Biometric application.BiometricService
}

// App is the fully wired backend application.
type App struct {
	Config   platform.Config
	Logger   *slog.Logger
	Pool     *pgxpool.Pool
	Echo     *echo.Echo
	Services AppServices
}

const bootstrapAdvisoryLockKey int64 = 901502001

var gooseDialectOnce sync.Once
var gooseDialectErr error

// Close releases resources owned by the app.
func (a *App) Close() {
	if a == nil || a.Pool == nil {
		return
	}
	a.Pool.Close()
}

// NewApp builds the production dependency graph and HTTP routes.
// When dumpFn is nil, pg_dump from PATH is used for backups.
func NewApp(ctx context.Context, cfg platform.Config, logger *slog.Logger, dumpFn application.DumpFunc) (*App, error) {
	if logger == nil {
		logger = platform.NewLogger(cfg.LogLevel)
	}
	if dumpFn == nil {
		dumpFn = defaultDumpFunc
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	poolCfg.MaxConns = 25
	poolCfg.MinConns = 5
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := ensureDatabaseBootstrap(ctx, cfg, logger); err != nil {
		pool.Close()
		return nil, err
	}

	userStore := postgres.NewUserStore(pool)
	sessStore := postgres.NewSessionStore(pool)
	captStore := postgres.NewCaptchaStore(pool)
	auditStore := postgres.NewAuditStore(pool)

	itemStore := postgres.NewItemStore(pool)
	availStore := postgres.NewAvailabilityWindowStore(pool)
	blackoutStore := postgres.NewBlackoutWindowStore(pool)
	inventoryStore := postgres.NewInventoryStore(pool)
	campaignStore := postgres.NewCampaignStore(pool)
	participantStore := postgres.NewParticipantStore(pool)
	orderStore := postgres.NewOrderStore(pool)
	timelineStore := postgres.NewTimelineStore(pool)

	supplierStore := postgres.NewSupplierStore(pool)
	poStore := postgres.NewPurchaseOrderStore(pool)
	poLineStore := postgres.NewPOLineStore(pool)
	varianceStore := postgres.NewVarianceStore(pool)
	landedCostStore := postgres.NewLandedCostStore(pool)
	reportStore := postgres.NewReportStore(pool)
	exportStore := postgres.NewExportStore(pool)
	backupStore := postgres.NewBackupStore(pool)
	retentionStore := postgres.NewRetentionStore(pool)
	biometricStore := postgres.NewBiometricStore(pool)
	encKeyStore := postgres.NewEncryptionKeyStore(pool)
	fulfillmentStore := postgres.NewFulfillmentStore(pool)

	locationStore := postgres.NewLocationStore(pool)
	memberStore := postgres.NewMemberStore(pool)
	coachStore := postgres.NewCoachStore(pool)

	auditSvc := application.NewAuditService(auditStore)
	authSvc := application.NewAuthService(userStore, sessStore, captStore, auditSvc, &cfg)
	itemSvc := application.NewItemService(itemStore, availStore, blackoutStore, itemStore, auditSvc, pool)
	inventorySvc := application.NewInventoryService(inventoryStore, inventoryStore, itemStore, auditSvc, pool)
	campaignSvc := application.NewCampaignService(campaignStore, participantStore, timelineStore, itemStore, availStore, blackoutStore, orderStore, inventoryStore, auditSvc, pool)
	orderSvc := application.NewOrderService(orderStore, timelineStore, itemStore, inventoryStore, availStore, blackoutStore, fulfillmentStore, auditSvc, pool)

	supplierSvc := application.NewSupplierService(supplierStore, auditSvc)
	poSvc := application.NewPurchaseOrderService(poStore, poLineStore, varianceStore, landedCostStore, inventoryStore, itemStore, auditSvc, pool)
	varianceSvc := application.NewVarianceService(varianceStore, poLineStore, itemStore, inventoryStore, auditSvc, pool)
	landedCostSvc := application.NewLandedCostService(landedCostStore)
	locationSvc := application.NewLocationService(locationStore)
	memberSvc := application.NewMemberService(memberStore)
	coachSvc := application.NewCoachService(coachStore)
	dashboardSvc := application.NewDashboardService(pool)
	backupSvc := application.NewBackupService(backupStore, cfg, dumpFn, auditSvc)
	retentionSvc := application.NewRetentionService(retentionStore, auditSvc, pool)
	biometricSvc := application.NewBiometricService(biometricStore, encKeyStore, auditSvc, pool, cfg.BiometricKeyRotationDays)
	userSvc := application.NewUserService(userStore, auditSvc)
	reportSvc := application.NewReportService(reportStore, exportStore, cfg, auditSvc, pool)

	authHandler := internalhttp.NewAuthHandler(authSvc, &cfg)
	authMW := internalhttp.NewAuthMiddleware(authSvc)
	itemH := internalhttp.NewItemHandler(itemSvc)
	inventoryH := internalhttp.NewInventoryHandler(inventorySvc)
	campaignH := internalhttp.NewCampaignHandler(campaignSvc)
	orderH := internalhttp.NewOrderHandler(orderSvc)
	supplierH := internalhttp.NewSupplierHandler(supplierSvc)
	procurementH := internalhttp.NewProcurementHandler(poSvc)
	varianceH := internalhttp.NewVarianceHandler(varianceSvc)
	reportH := internalhttp.NewReportHandler(reportSvc)
	backupH := internalhttp.NewBackupHandler(backupSvc)
	retentionH := internalhttp.NewRetentionHandler(retentionSvc)
	biometricH := internalhttp.NewBiometricHandler(biometricSvc, cfg)
	userH := internalhttp.NewUserHandler(userSvc)
	adminH := internalhttp.NewAdminHandler(auditSvc)
	dashboardH := internalhttp.NewDashboardHandler(dashboardSvc)
	locationH := internalhttp.NewLocationHandler(locationSvc)
	memberH := internalhttp.NewMemberHandler(memberSvc)
	coachH := internalhttp.NewCoachHandler(coachSvc)
	landedCostH := internalhttp.NewLandedCostHandler(landedCostSvc)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = internalhttp.NewCustomValidator()
	e.Use(platform.LoggerMiddleware(logger))

	internalhttp.RegisterRoutes(
		e, logger, pool,
		authHandler, authMW,
		itemH, inventoryH, campaignH, orderH,
		supplierH, procurementH, varianceH,
		reportH, backupH, retentionH, biometricH,
		userH, adminH,
		dashboardH, locationH, memberH, coachH, landedCostH,
	)

	return &App{
		Config: cfg,
		Logger: logger,
		Pool:   pool,
		Echo:   e,
		Services: AppServices{
			Campaign:  campaignSvc,
			Order:     orderSvc,
			Backup:    backupSvc,
			Retention: retentionSvc,
			Variance:  varianceSvc,
			Biometric: biometricSvc,
		},
	}, nil
}

func defaultDumpFunc(ctx context.Context, connStr, destPath string) error {
	cmd := exec.CommandContext(ctx, "pg_dump",
		"--dbname="+connStr,
		"--file="+destPath,
		"--format=custom",
		"--no-password",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pg_dump failed: %w\noutput: %s", err, string(out))
	}
	return nil
}

func ensureDatabaseBootstrap(ctx context.Context, cfg platform.Config, logger *slog.Logger) error {
	if !cfg.RunMigrationsOnStartup && !cfg.RunSeedOnStartup {
		logger.Info("database bootstrap skipped", "migrations", false, "seed", false)
		return nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open bootstrap database connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping bootstrap database connection: %w", err)
	}

	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, bootstrapAdvisoryLockKey); err != nil {
		return fmt.Errorf("acquire bootstrap advisory lock: %w", err)
	}
	defer func() {
		if _, unlockErr := db.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, bootstrapAdvisoryLockKey); unlockErr != nil {
			logger.Warn("failed to release bootstrap advisory lock", "error", unlockErr)
		}
	}()

	if cfg.RunMigrationsOnStartup {
		gooseDialectOnce.Do(func() {
			gooseDialectErr = goose.SetDialect("postgres")
		})
		if gooseDialectErr != nil {
			return fmt.Errorf("set goose dialect: %w", gooseDialectErr)
		}

		migrationsDir := filepath.Join("database", "migrations")
		if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
			return fmt.Errorf("run goose migrations from %s: %w", migrationsDir, err)
		}
		logger.Info("database migrations applied", "dir", migrationsDir)
	}

	if cfg.RunSeedOnStartup {
		seedPath := filepath.Join("database", "seeds", "seed.sql")
		seedSQL, err := os.ReadFile(seedPath)
		if err != nil {
			return fmt.Errorf("read seed file %s: %w", seedPath, err)
		}
		if _, err := db.ExecContext(ctx, string(seedSQL)); err != nil {
			return fmt.Errorf("execute seed file %s: %w", seedPath, err)
		}
		logger.Info("seed data applied", "path", seedPath)
	}

	return nil
}
