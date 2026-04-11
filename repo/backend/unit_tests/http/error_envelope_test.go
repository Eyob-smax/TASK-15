package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"fitcommerce/internal/domain"
	fchttp "fitcommerce/internal/http"
)

func newTestContext() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func parseErrorBody(t *testing.T, rec *httptest.ResponseRecorder) fchttp.ErrorResponse {
	t.Helper()
	var resp fchttp.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v\nbody: %s", err, rec.Body.String())
	}
	return resp
}

func TestHandleDomainError_NotFound(t *testing.T) {
	c, rec := newTestContext()
	_ = fchttp.HandleDomainError(c, domain.ErrNotFound)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	body := parseErrorBody(t, rec)
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", body.Error.Code)
	}
}

func TestHandleDomainError_Conflict(t *testing.T) {
	c, rec := newTestContext()
	_ = fchttp.HandleDomainError(c, &domain.ErrConflict{Entity: "item", Message: "version mismatch"})
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestHandleDomainError_Validation(t *testing.T) {
	c, rec := newTestContext()
	_ = fchttp.HandleDomainError(c, &domain.ErrValidation{Field: "email", Message: "invalid"})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestHandleDomainError_AccountLocked(t *testing.T) {
	c, rec := newTestContext()
	_ = fchttp.HandleDomainError(c, &domain.ErrAccountLocked{Message: "too many attempts"})
	if rec.Code != http.StatusLocked {
		t.Errorf("expected 423, got %d", rec.Code)
	}
}

func TestHandleDomainError_CaptchaRequired(t *testing.T) {
	c, rec := newTestContext()
	_ = fchttp.HandleDomainError(c, &domain.ErrCaptchaRequired{ChallengeID: "abc123"})
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandleDomainError_UnknownError(t *testing.T) {
	c, rec := newTestContext()
	_ = fchttp.HandleDomainError(c, errors.New("something unexpected"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}
