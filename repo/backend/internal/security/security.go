// Package security provides cryptographic and access-control utilities
// for the FitCommerce platform.
//
// Implemented components:
//
//   - Argon2id password hashing (Memory=64MB, Iterations=3, Parallelism=2,
//     KeyLen=32, SaltLen=16) with constant-time verification (password.go).
//
//   - On-device math-based CAPTCHA challenge generation and answer verification
//     for login lockout flows; no external service required (captcha.go).
//
//   - Session token generation, cookie management (HttpOnly, SameSite=Strict),
//     and token extraction from Echo contexts (session.go).
//
//   - RBAC helpers: 15 permission action constants, per-role permission matrix,
//     and Echo context helpers for storing/retrieving the authenticated user (rbac.go).
//
//   - Data masking utilities for PII fields (email, phone) with role-aware
//     policy enforcement and in-place UserResponse masking (masking.go).
//
//   - AES-256-GCM encryption and decryption for biometric data at rest;
//     key derivation from operator-configured references is deferred to Prompt 7 (crypto.go).
//
//   - Audit helper: event type constants, tamper-evident event construction with
//     SHA-256 hash chaining, and sensitive-field redaction (audit_helper.go).
package security
