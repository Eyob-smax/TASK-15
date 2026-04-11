package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"fitcommerce/internal/application"
	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
	"fitcommerce/internal/platform"
	"fitcommerce/internal/security"
)

// BiometricHandler handles HTTP requests for biometric and encryption key endpoints.
type BiometricHandler struct {
	svc application.BiometricService
	cfg platform.Config
}

// NewBiometricHandler creates a BiometricHandler backed by the given service and config.
func NewBiometricHandler(svc application.BiometricService, cfg platform.Config) *BiometricHandler {
	return &BiometricHandler{svc: svc, cfg: cfg}
}

// RegisterBiometric handles POST /admin/biometrics.
func (h *BiometricHandler) RegisterBiometric(c echo.Context) error {
	if !h.cfg.BiometricModuleEnabled {
		return c.JSON(http.StatusNotImplemented, NewErrorResponse("MODULE_DISABLED", "biometric module is not enabled"))
	}

	var req dto.RegisterBiometricRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid request body"))
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid user_id"))
	}

	enrollment, err := h.svc.Register(c.Request().Context(), userID, req.TemplateRef)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusCreated, dto.SuccessResponse{Data: toBiometricResponse(enrollment)})
}

// GetBiometric handles GET /admin/biometrics/:user_id.
func (h *BiometricHandler) GetBiometric(c echo.Context) error {
	if !h.cfg.BiometricModuleEnabled {
		return c.JSON(http.StatusNotImplemented, NewErrorResponse("MODULE_DISABLED", "biometric module is not enabled"))
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid user_id"))
	}

	enrollment, err := h.svc.GetByUser(c.Request().Context(), userID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toBiometricResponse(enrollment)})
}

// RevokeBiometric handles POST /admin/biometrics/:user_id/revoke.
func (h *BiometricHandler) RevokeBiometric(c echo.Context) error {
	if !h.cfg.BiometricModuleEnabled {
		return c.JSON(http.StatusNotImplemented, NewErrorResponse("MODULE_DISABLED", "biometric module is not enabled"))
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse("BAD_REQUEST", "invalid user_id"))
	}

	if err := h.svc.Revoke(c.Request().Context(), userID); err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: map[string]string{"message": "biometric enrollment revoked"}})
}

// RotateKey handles POST /admin/biometrics/rotate-key.
func (h *BiometricHandler) RotateKey(c echo.Context) error {
	if !h.cfg.BiometricModuleEnabled {
		return c.JSON(http.StatusNotImplemented, NewErrorResponse("MODULE_DISABLED", "biometric module is not enabled"))
	}

	user, ok := security.GetUserFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, NewErrorResponse("UNAUTHORIZED", "not authenticated"))
	}

	key, err := h.svc.RotateKey(c.Request().Context(), user.ID)
	if err != nil {
		return HandleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: toEncryptionKeyResponse(key)})
}

// ListKeys handles GET /admin/biometrics/keys.
func (h *BiometricHandler) ListKeys(c echo.Context) error {
	if !h.cfg.BiometricModuleEnabled {
		return c.JSON(http.StatusNotImplemented, NewErrorResponse("MODULE_DISABLED", "biometric module is not enabled"))
	}

	keys, err := h.svc.ListKeys(c.Request().Context())
	if err != nil {
		return HandleDomainError(c, err)
	}

	resp := make([]dto.EncryptionKeyResponse, len(keys))
	for i := range keys {
		resp[i] = toEncryptionKeyResponse(&keys[i])
	}
	return c.JSON(http.StatusOK, dto.SuccessResponse{Data: resp})
}

// --- Private helpers ---

func toBiometricResponse(e *domain.BiometricEnrollment) dto.BiometricEnrollmentResponse {
	return dto.BiometricEnrollmentResponse{
		ID:          e.ID.String(),
		UserID:      e.UserID.String(),
		TemplateRef: "[BIOMETRIC REDACTED]",
		IsActive:    len(e.EncryptedData) > 0,
		CreatedAt:   e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   e.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toEncryptionKeyResponse(k *domain.EncryptionKey) dto.EncryptionKeyResponse {
	return dto.EncryptionKeyResponse{
		ID:        k.ID.String(),
		Purpose:   k.Purpose,
		IsActive:  k.Status == domain.EncryptionKeyStatusActive,
		CreatedAt: k.ActivatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
