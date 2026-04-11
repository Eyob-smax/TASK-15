package security

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"fitcommerce/internal/domain"
	"fitcommerce/internal/http/dto"
)

// MaskEmail returns a masked version of the email address, revealing only the
// first character before the @ and the full domain portion.
// Example: "alice@example.com" → "a***@example.com"
func MaskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***@***"
	}
	return fmt.Sprintf("%s***@%s", string(parts[0][0]), parts[1])
}

// MaskPhone returns a masked version of the phone number, revealing only the
// last four digits. All other characters are replaced with *.
// Example: "+1-555-867-5309" → "****5309"
func MaskPhone(phone string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	if len(digits) <= 4 {
		return strings.Repeat("*", len(digits))
	}
	return fmt.Sprintf("%s%s", strings.Repeat("*", len(digits)-4), digits[len(digits)-4:])
}

// RedactBiometric returns a constant redaction placeholder for any biometric
// data field, ensuring raw biometric values never appear in responses or logs.
func RedactBiometric() string {
	return "[BIOMETRIC REDACTED]"
}

// MaskFieldByRole applies the masking policy for a specific field based on the
// requesting user's role and whether they are viewing their own record.
// ownerID is the UUID of the entity being viewed; requesterID is the caller.
func MaskFieldByRole(fieldName, value string, role domain.UserRole, ownerID, requesterID uuid.UUID) string {
	isOwn := ownerID == requesterID

	switch fieldName {
	case "email":
		switch role {
		case domain.UserRoleAdministrator, domain.UserRoleOperationsManager, domain.UserRoleProcurementSpecialist:
			return value
		case domain.UserRoleCoach:
			if isOwn {
				return value
			}
			return MaskEmail(value)
		case domain.UserRoleMember:
			if isOwn {
				return value
			}
			return MaskEmail(value)
		}

	case "phone":
		switch role {
		case domain.UserRoleAdministrator:
			return value
		case domain.UserRoleOperationsManager:
			return MaskPhone(value)
		default:
			if isOwn {
				return MaskPhone(value)
			}
			return ""
		}

	case "biometric", "encrypted_data":
		return RedactBiometric()

	case "password_hash", "salt", "session_token", "token":
		return ""
	}

	return value
}

// ApplyUserResponseMask applies in-place masking to a UserResponse DTO based on
// the requesting role and whether the requester is viewing their own record.
// Fields that must never be returned (password_hash, salt) are already excluded
// from UserResponse by construction.
func ApplyUserResponseMask(resp *dto.UserResponse, role domain.UserRole, requesterID uuid.UUID) {
	if resp == nil {
		return
	}

	ownerID, err := uuid.Parse(resp.ID)
	if err != nil {
		ownerID = uuid.Nil
	}

	resp.Email = MaskFieldByRole("email", resp.Email, role, ownerID, requesterID)
}
