package platform_test

import (
	"os"
	"testing"

	"fitcommerce/internal/platform"
)

// setenv sets an environment variable and registers a cleanup that restores the
// original value (or unsets it if it was not originally set).
func setenv(t *testing.T, key, value string) {
	t.Helper()
	original, had := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if had {
			os.Setenv(key, original)
		} else {
			os.Unsetenv(key)
		}
	})
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Unset all FC_ vars to ensure defaults are applied.
	for _, key := range []string{
		"FC_SERVER_PORT", "FC_DATABASE_URL", "FC_SESSION_IDLE_TIMEOUT_MINUTES",
		"FC_SESSION_ABSOLUTE_TIMEOUT_HOURS", "FC_LOGIN_LOCKOUT_THRESHOLD",
		"FC_LOGIN_LOCKOUT_DURATION_MINUTES", "FC_BACKUP_PATH",
		"FC_BACKUP_ENCRYPTION_KEY_REF", "FC_EXPORT_PATH",
		"FC_TLS_CERT_FILE", "FC_TLS_KEY_FILE",
		"FC_ALLOW_INSECURE_HTTP",
		"FC_RUN_MIGRATIONS_ON_STARTUP", "FC_RUN_SEED_ON_STARTUP",
		"FC_BIOMETRIC_MODULE_ENABLED", "FC_BIOMETRIC_KEY_ROTATION_DAYS",
		"FC_CLUB_TIMEZONE", "FC_LOG_LEVEL",
	} {
		original, had := os.LookupEnv(key)
		os.Unsetenv(key)
		keyCopy := key
		if had {
			originalCopy := original
			t.Cleanup(func() { os.Setenv(keyCopy, originalCopy) })
		} else {
			t.Cleanup(func() { os.Unsetenv(keyCopy) })
		}
	}

	cfg := platform.LoadConfig()

	if cfg.ServerPort != "8080" {
		t.Errorf("ServerPort default: want 8080, got %q", cfg.ServerPort)
	}
	if cfg.SessionIdleTimeoutMinutes != 30 {
		t.Errorf("SessionIdleTimeoutMinutes default: want 30, got %d", cfg.SessionIdleTimeoutMinutes)
	}
	if cfg.SessionAbsoluteTimeoutHours != 12 {
		t.Errorf("SessionAbsoluteTimeoutHours default: want 12, got %d", cfg.SessionAbsoluteTimeoutHours)
	}
	if cfg.LoginLockoutThreshold != 5 {
		t.Errorf("LoginLockoutThreshold default: want 5, got %d", cfg.LoginLockoutThreshold)
	}
	if cfg.LoginLockoutDurationMinutes != 15 {
		t.Errorf("LoginLockoutDurationMinutes default: want 15, got %d", cfg.LoginLockoutDurationMinutes)
	}
	if cfg.BackupPath != "/var/backups/fitcommerce" {
		t.Errorf("BackupPath default: want /var/backups/fitcommerce, got %q", cfg.BackupPath)
	}
	if cfg.BackupEncryptionKeyRef != "" {
		t.Errorf("BackupEncryptionKeyRef default: want empty, got %q", cfg.BackupEncryptionKeyRef)
	}
	if cfg.ExportPath != "/tmp/fitcommerce-exports" {
		t.Errorf("ExportPath default: want /tmp/fitcommerce-exports, got %q", cfg.ExportPath)
	}
	if cfg.AllowInsecureHTTP {
		t.Error("AllowInsecureHTTP default: want false")
	}
	if !cfg.RunMigrationsOnStartup {
		t.Error("RunMigrationsOnStartup default: want true")
	}
	if !cfg.RunSeedOnStartup {
		t.Error("RunSeedOnStartup default: want true")
	}
	if cfg.BiometricModuleEnabled != false {
		t.Errorf("BiometricModuleEnabled default: want false, got %v", cfg.BiometricModuleEnabled)
	}
	if cfg.BiometricKeyRotationDays != 90 {
		t.Errorf("BiometricKeyRotationDays default: want 90, got %d", cfg.BiometricKeyRotationDays)
	}
	if cfg.ClubTimezone != "UTC" {
		t.Errorf("ClubTimezone default: want UTC, got %q", cfg.ClubTimezone)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel default: want info, got %q", cfg.LogLevel)
	}
}

func TestLoadConfig_OverridesFromEnv(t *testing.T) {
	setenv(t, "FC_SERVER_PORT", "9090")
	setenv(t, "FC_LOG_LEVEL", "debug")
	setenv(t, "FC_ALLOW_INSECURE_HTTP", "true")
	setenv(t, "FC_BIOMETRIC_MODULE_ENABLED", "true")
	setenv(t, "FC_SESSION_IDLE_TIMEOUT_MINUTES", "60")
	setenv(t, "FC_CLUB_TIMEZONE", "America/Chicago")
	setenv(t, "FC_BACKUP_ENCRYPTION_KEY_REF", "secret-ref")
	setenv(t, "FC_EXPORT_PATH", "/data/exports")

	cfg := platform.LoadConfig()

	if cfg.ServerPort != "9090" {
		t.Errorf("ServerPort env override: want 9090, got %q", cfg.ServerPort)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel env override: want debug, got %q", cfg.LogLevel)
	}
	if !cfg.AllowInsecureHTTP {
		t.Error("AllowInsecureHTTP env override: want true")
	}
	if !cfg.BiometricModuleEnabled {
		t.Error("BiometricModuleEnabled env override: want true")
	}
	if cfg.SessionIdleTimeoutMinutes != 60 {
		t.Errorf("SessionIdleTimeoutMinutes env override: want 60, got %d", cfg.SessionIdleTimeoutMinutes)
	}
	if cfg.ClubTimezone != "America/Chicago" {
		t.Errorf("ClubTimezone env override: want America/Chicago, got %q", cfg.ClubTimezone)
	}
	if cfg.BackupEncryptionKeyRef != "secret-ref" {
		t.Errorf("BackupEncryptionKeyRef env override: want secret-ref, got %q", cfg.BackupEncryptionKeyRef)
	}
	if cfg.ExportPath != "/data/exports" {
		t.Errorf("ExportPath env override: want /data/exports, got %q", cfg.ExportPath)
	}
}

func TestLoadConfig_InvalidInt_FallsBackToDefault(t *testing.T) {
	setenv(t, "FC_SESSION_IDLE_TIMEOUT_MINUTES", "not-a-number")

	cfg := platform.LoadConfig()

	// Should use the default (30) when the value is not parseable.
	if cfg.SessionIdleTimeoutMinutes != 30 {
		t.Errorf("expected fallback to default 30 for invalid int, got %d", cfg.SessionIdleTimeoutMinutes)
	}
}

func TestLoadConfig_InvalidBool_FallsBackToDefault(t *testing.T) {
	setenv(t, "FC_BIOMETRIC_MODULE_ENABLED", "yes-please")

	cfg := platform.LoadConfig()

	// Should use the default (false) when the value is not parseable.
	if cfg.BiometricModuleEnabled != false {
		t.Errorf("expected fallback to default false for invalid bool, got %v", cfg.BiometricModuleEnabled)
	}
}

func TestLoadConfig_StartupBootstrapFlags(t *testing.T) {
	setenv(t, "FC_RUN_MIGRATIONS_ON_STARTUP", "false")
	setenv(t, "FC_RUN_SEED_ON_STARTUP", "false")

	cfg := platform.LoadConfig()

	if cfg.RunMigrationsOnStartup {
		t.Error("RunMigrationsOnStartup env override: want false")
	}
	if cfg.RunSeedOnStartup {
		t.Error("RunSeedOnStartup env override: want false")
	}
}

func TestLoadConfig_DatabaseURL_Default(t *testing.T) {
	original, had := os.LookupEnv("FC_DATABASE_URL")
	os.Unsetenv("FC_DATABASE_URL")
	defer func() {
		if had {
			os.Setenv("FC_DATABASE_URL", original)
		} else {
			os.Unsetenv("FC_DATABASE_URL")
		}
	}()

	cfg := platform.LoadConfig()

	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL should have a non-empty default")
	}
}
