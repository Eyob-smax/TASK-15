package platform

import (
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
// All environment variables use the FC_ prefix.
type Config struct {
	ServerPort                   string
	DatabaseURL                  string
	RunMigrationsOnStartup       bool
	RunSeedOnStartup             bool
	SessionIdleTimeoutMinutes    int
	SessionAbsoluteTimeoutHours  int
	LoginLockoutThreshold        int
	LoginLockoutDurationMinutes  int
	BackupPath                   string
	BackupEncryptionKeyRef       string
	ExportPath                   string
	TLSCertFile                  string
	TLSKeyFile                   string
	AllowInsecureHTTP            bool
	BiometricModuleEnabled       bool
	BiometricKeyRotationDays     int
	ClubTimezone                 string
	LogLevel                     string
}

// LoadConfig reads configuration from environment variables with FC_ prefix,
// applying defaults where appropriate.
func LoadConfig() Config {
	return Config{
		ServerPort:                   envOrDefault("FC_SERVER_PORT", "8080"),
		DatabaseURL:                  envOrDefault("FC_DATABASE_URL", "postgres://fitcommerce:fitcommerce@localhost:5432/fitcommerce?sslmode=disable"),
		RunMigrationsOnStartup:       envOrDefaultBool("FC_RUN_MIGRATIONS_ON_STARTUP", true),
		RunSeedOnStartup:             envOrDefaultBool("FC_RUN_SEED_ON_STARTUP", true),
		SessionIdleTimeoutMinutes:    envOrDefaultInt("FC_SESSION_IDLE_TIMEOUT_MINUTES", 30),
		SessionAbsoluteTimeoutHours:  envOrDefaultInt("FC_SESSION_ABSOLUTE_TIMEOUT_HOURS", 12),
		LoginLockoutThreshold:        envOrDefaultInt("FC_LOGIN_LOCKOUT_THRESHOLD", 5),
		LoginLockoutDurationMinutes:  envOrDefaultInt("FC_LOGIN_LOCKOUT_DURATION_MINUTES", 15),
		BackupPath:                   envOrDefault("FC_BACKUP_PATH", "/var/backups/fitcommerce"),
		BackupEncryptionKeyRef:       envOrDefault("FC_BACKUP_ENCRYPTION_KEY_REF", ""),
		ExportPath:                   envOrDefault("FC_EXPORT_PATH", "/tmp/fitcommerce-exports"),
		TLSCertFile:                  envOrDefault("FC_TLS_CERT_FILE", ""),
		TLSKeyFile:                   envOrDefault("FC_TLS_KEY_FILE", ""),
		AllowInsecureHTTP:            envOrDefaultBool("FC_ALLOW_INSECURE_HTTP", false),
		BiometricModuleEnabled:       envOrDefaultBool("FC_BIOMETRIC_MODULE_ENABLED", false),
		BiometricKeyRotationDays:     envOrDefaultInt("FC_BIOMETRIC_KEY_ROTATION_DAYS", 90),
		ClubTimezone:                 envOrDefault("FC_CLUB_TIMEZONE", "UTC"),
		LogLevel:                     envOrDefault("FC_LOG_LEVEL", "info"),
	}
}

func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func envOrDefaultBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return parsed
}
