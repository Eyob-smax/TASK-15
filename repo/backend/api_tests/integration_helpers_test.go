package api_tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"fitcommerce/internal/bootstrap"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/security"
)

const defaultTestPassword = "Password123!"

var (
	apiDBMu   sync.Mutex
	gooseOnce sync.Once
)

type integrationApp struct {
	app    *bootstrap.App
	cfg    platform.Config
	server *httptest.Server
	client *http.Client
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"details"`
	} `json:"error"`
}

type userSeed struct {
	ID       uuid.UUID
	Email    string
	Password string
	Role     domain.UserRole
}

type locationSeed struct {
	ID   uuid.UUID
	Name string
}

type itemSeed struct {
	ID       uuid.UUID
	Name     string
	Version  int
	Quantity int
}

type timeWindow struct {
	Start time.Time
	End   time.Time
}

type itemSeedOptions struct {
	CreatedBy           uuid.UUID
	Name                string
	Category            string
	Brand               string
	Condition           domain.ItemCondition
	BillingModel        domain.BillingModel
	Status              domain.ItemStatus
	Quantity            int
	UnitPrice           float64
	RefundableDeposit   float64
	LocationID          *uuid.UUID
	AvailabilityWindows []timeWindow
	BlackoutWindows     []timeWindow
}

func newIntegrationApp(t *testing.T) *integrationApp {
	t.Helper()
	return newIntegrationAppWithConfig(t, integrationConfig(t))
}

func newIntegrationAppWithConfig(t *testing.T, cfg platform.Config) *integrationApp {
	t.Helper()

	apiDBMu.Lock()
	t.Cleanup(func() {
		apiDBMu.Unlock()
	})

	resetDatabase(t, cfg)

	testDump := func(ctx context.Context, connStr, destPath string) error {
		_ = ctx
		_ = connStr
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(destPath, []byte("integration-test-backup"), 0600)
	}

	app, err := bootstrap.NewApp(context.Background(), cfg, platform.NewLogger("error"), testDump)
	if err != nil {
		t.Fatalf("bootstrap app: %v", err)
	}

	// Start a real HTTP test server so all API calls go through the full HTTP
	// stack (middleware, routing, serialisation) over an actual TCP connection.
	ts := httptest.NewServer(app.Echo)
	t.Cleanup(ts.Close)
	t.Cleanup(func() {
		app.Close()
	})

	// Disable automatic redirect following so tests receive the raw response
	// status exactly as the server sends it.
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &integrationApp{
		app:    app,
		cfg:    cfg,
		server: ts,
		client: client,
	}
}

func integrationConfig(t *testing.T) platform.Config {
	t.Helper()

	exportDir := filepath.Join(t.TempDir(), "exports")
	backupDir := filepath.Join(t.TempDir(), "backups")

	return platform.Config{
		ServerPort:                  "8080",
		DatabaseURL:                 envOr("FC_DATABASE_URL", "postgres://fitcommerce:fitcommerce@postgres:5432/fitcommerce?sslmode=disable"),
		SessionIdleTimeoutMinutes:   30,
		SessionAbsoluteTimeoutHours: 12,
		LoginLockoutThreshold:       5,
		LoginLockoutDurationMinutes: 15,
		BackupPath:                  backupDir,
		BackupEncryptionKeyRef:      "integration-test-backup-key",
		ExportPath:                  exportDir,
		AllowInsecureHTTP:           true,
		RunMigrationsOnStartup:      false,
		RunSeedOnStartup:            false,
		BiometricModuleEnabled:      false,
		BiometricKeyRotationDays:    90,
		ClubTimezone:                "UTC",
		LogLevel:                    "error",
	}
}

func resetDatabase(t *testing.T, cfg platform.Config) {
	t.Helper()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	statements := []string{
		`DROP SCHEMA IF EXISTS public CASCADE`,
		`CREATE SCHEMA public`,
		`CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("exec %q: %v", stmt, err)
		}
	}

	if _, err := db.ExecContext(ctx, `SET search_path TO public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	gooseOnce.Do(func() {
		if err := goose.SetDialect("postgres"); err != nil {
			panic(err)
		}
	})

	if err := goose.UpContext(ctx, db, filepath.Join(backendRoot(t), "database", "migrations")); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	seedSQL, err := os.ReadFile(filepath.Join(backendRoot(t), "database", "seeds", "seed.sql"))
	if err != nil {
		t.Fatalf("read seed.sql: %v", err)
	}
	if _, err := db.ExecContext(ctx, string(seedSQL)); err != nil {
		t.Fatalf("run seed.sql: %v", err)
	}
}

func backendRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Dir(filepath.Dir(file))
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func (a *integrationApp) request(t *testing.T, method, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, a.server.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	// Wrap the real HTTP response in a ResponseRecorder so all existing
	// assertions (rec.Code, rec.Body.Bytes(), rec.Header()) stay unchanged.
	return &httptest.ResponseRecorder{
		Code:      resp.StatusCode,
		HeaderMap: resp.Header,
		Body:      bytes.NewBuffer(respBody),
	}
}

func (a *integrationApp) get(t *testing.T, path string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return a.request(t, http.MethodGet, path, nil, cookies)
}

func (a *integrationApp) post(t *testing.T, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return a.request(t, http.MethodPost, path, body, cookies)
}

func (a *integrationApp) put(t *testing.T, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return a.request(t, http.MethodPut, path, body, cookies)
}

func responseCookies(rec *httptest.ResponseRecorder) []*http.Cookie {
	resp := &http.Response{Header: rec.Header()}
	return resp.Cookies()
}

func decodeSuccess[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var envelope struct {
		Data T `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode success body: %v\nbody=%s", err, rec.Body.String())
	}
	return envelope.Data
}

func decodePaginated[T any](t *testing.T, rec *httptest.ResponseRecorder) ([]T, dto.PaginationMeta) {
	t.Helper()

	var envelope struct {
		Data       []T                `json:"data"`
		Pagination dto.PaginationMeta `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode paginated body: %v\nbody=%s", err, rec.Body.String())
	}
	return envelope.Data, envelope.Pagination
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) errorEnvelope {
	t.Helper()

	var envelope errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error body: %v\nbody=%s", err, rec.Body.String())
	}
	return envelope
}

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("expected status %d, got %d: %s", want, rec.Code, rec.Body.String())
	}
}

func (a *integrationApp) seedLocation(t *testing.T, name string) locationSeed {
	t.Helper()

	id := uuid.New()
	if name == "" {
		name = "Location " + shortID()
	}

	_, err := a.app.Pool.Exec(context.Background(),
		`INSERT INTO locations (id, name, address, timezone, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
		id, name, "123 Training Ave", "UTC",
	)
	if err != nil {
		t.Fatalf("seed location: %v", err)
	}

	return locationSeed{ID: id, Name: name}
}

func (a *integrationApp) seedUser(t *testing.T, role domain.UserRole, locationID *uuid.UUID) userSeed {
	t.Helper()

	id := uuid.New()
	email := fmt.Sprintf("%s-%s@fitcommerce.test", strings.ReplaceAll(string(role), "_", "-"), shortID())
	displayName := strings.Title(strings.ReplaceAll(string(role), "_", " "))
	hash, salt, err := security.HashPassword(defaultTestPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	_, err = a.app.Pool.Exec(context.Background(),
		`INSERT INTO users (
			id, email, password_hash, salt, role, status, display_name, location_id,
			failed_login_count, locked_until, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, 'active', $6, $7, 0, NULL, NOW(), NOW())`,
		id, email, hash, salt, role, displayName, locationID,
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return userSeed{
		ID:       id,
		Email:    email,
		Password: defaultTestPassword,
		Role:     role,
	}
}

func (a *integrationApp) seedMemberRecord(t *testing.T, userID, locationID uuid.UUID, status domain.MembershipStatus, joinedAt, updatedAt time.Time) uuid.UUID {
	t.Helper()

	id := uuid.New()
	_, err := a.app.Pool.Exec(context.Background(),
		`INSERT INTO members (
			id, user_id, location_id, membership_status, joined_at, renewal_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7)`,
		id, userID, locationID, status, joinedAt, joinedAt.AddDate(0, 1, 0), updatedAt,
	)
	if err != nil {
		t.Fatalf("seed member record: %v", err)
	}
	return id
}

func (a *integrationApp) seedCoachRecord(t *testing.T, userID, locationID uuid.UUID, specialization string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	_, err := a.app.Pool.Exec(context.Background(),
		`INSERT INTO coaches (
			id, user_id, location_id, specialization, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, true, NOW(), NOW())`,
		id, userID, locationID, specialization,
	)
	if err != nil {
		t.Fatalf("seed coach record: %v", err)
	}
	return id
}

func (a *integrationApp) seedItem(t *testing.T, opts itemSeedOptions) itemSeed {
	t.Helper()

	if opts.Name == "" {
		opts.Name = "Item " + shortID()
	}
	if opts.Category == "" {
		opts.Category = "strength"
	}
	if opts.Brand == "" {
		opts.Brand = "FitCommerce"
	}
	if opts.Condition == "" {
		opts.Condition = domain.ItemConditionNew
	}
	if opts.BillingModel == "" {
		opts.BillingModel = domain.BillingModelOneTime
	}
	if opts.Status == "" {
		opts.Status = domain.ItemStatusDraft
	}
	if opts.Quantity == 0 {
		opts.Quantity = 10
	}
	if opts.RefundableDeposit == 0 {
		opts.RefundableDeposit = 50
	}

	id := uuid.New()
	_, err := a.app.Pool.Exec(context.Background(),
		`INSERT INTO items (
			id, sku, name, description, category, brand, condition, billing_model,
			unit_price, refundable_deposit, quantity, status, location_id, created_by,
			version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14,
			1, NOW(), NOW()
		)`,
		id,
		"SKU-"+strings.ToUpper(shortID()),
		opts.Name,
		"Seeded test item",
		opts.Category,
		opts.Brand,
		opts.Condition,
		opts.BillingModel,
		opts.UnitPrice,
		opts.RefundableDeposit,
		opts.Quantity,
		opts.Status,
		opts.LocationID,
		opts.CreatedBy,
	)
	if err != nil {
		t.Fatalf("seed item: %v", err)
	}

	for _, window := range opts.AvailabilityWindows {
		_, err := a.app.Pool.Exec(context.Background(),
			`INSERT INTO item_availability_windows (id, item_id, start_time, end_time)
			 VALUES ($1, $2, $3, $4)`,
			uuid.New(), id, window.Start, window.End,
		)
		if err != nil {
			t.Fatalf("seed availability window: %v", err)
		}
	}

	for _, window := range opts.BlackoutWindows {
		_, err := a.app.Pool.Exec(context.Background(),
			`INSERT INTO item_blackout_windows (id, item_id, start_time, end_time)
			 VALUES ($1, $2, $3, $4)`,
			uuid.New(), id, window.Start, window.End,
		)
		if err != nil {
			t.Fatalf("seed blackout window: %v", err)
		}
	}

	return itemSeed{
		ID:       id,
		Name:     opts.Name,
		Version:  1,
		Quantity: opts.Quantity,
	}
}

func (a *integrationApp) seedSupplier(t *testing.T, name string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	if name == "" {
		name = "Supplier " + shortID()
	}

	_, err := a.app.Pool.Exec(context.Background(),
		`INSERT INTO suppliers (
			id, name, contact_name, contact_email, contact_phone, address, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())`,
		id, name, "Jordan Vendor", "vendor@fitcommerce.test", "555-1000", "Supply Street",
	)
	if err != nil {
		t.Fatalf("seed supplier: %v", err)
	}
	return id
}

func (a *integrationApp) login(t *testing.T, user userSeed) []*http.Cookie {
	t.Helper()

	rec := a.post(t, "/api/v1/auth/login", map[string]string{
		"email":    user.Email,
		"password": user.Password,
	}, nil)
	requireStatus(t, rec, http.StatusOK)
	return responseCookies(rec)
}

func (a *integrationApp) reportIDByType(t *testing.T, reportType string) uuid.UUID {
	t.Helper()

	// Seed a temporary admin to authenticate the list request. Report definitions
	// are seeded by seed.sql and are always present after resetDatabase.
	admin := a.seedUser(t, "administrator", nil)
	rec := a.get(t, "/api/v1/reports", a.login(t, admin))
	requireStatus(t, rec, http.StatusOK)

	reports := decodeSuccess[[]dto.ReportResponse](t, rec)
	for _, r := range reports {
		if r.ReportType == reportType {
			id, err := uuid.Parse(r.ID)
			if err != nil {
				t.Fatalf("parse report id %q: %v", r.ID, err)
			}
			return id
		}
	}
	t.Fatalf("report type %q not found via GET /api/v1/reports", reportType)
	return uuid.Nil
}

func (a *integrationApp) captchaAnswer(t *testing.T, challengeID string) string {
	t.Helper()

	var challenge string
	err := a.app.Pool.QueryRow(context.Background(),
		`SELECT challenge_data FROM captcha_challenges WHERE id = $1`,
		challengeID,
	).Scan(&challenge)
	if err != nil {
		t.Fatalf("query captcha challenge: %v", err)
	}

	var x int
	var y int
	if _, err := fmt.Sscanf(challenge, "What is %d + %d?", &x, &y); err != nil {
		t.Fatalf("parse captcha challenge %q: %v", challenge, err)
	}
	return fmt.Sprintf("%d", x+y)
}

func shortID() string {
	return strings.ReplaceAll(uuid.NewString()[:8], "-", "")
}
