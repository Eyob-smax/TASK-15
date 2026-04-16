package platform

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// EnsureServerTLSFiles returns the configured TLS files when both paths are
// provided. When no explicit files are configured it generates an ephemeral
// self-signed certificate so the server can still enforce encrypted transport
// on local networks and single-machine deployments.
func EnsureServerTLSFiles(certFile, keyFile string) (resolvedCertFile, resolvedKeyFile string, cleanup func(), err error) {
	if certFile != "" || keyFile != "" {
		if certFile == "" || keyFile == "" {
			return "", "", nil, fmt.Errorf("both certificate and key files must be provided together")
		}
		if filesExist(certFile, keyFile) {
			return certFile, keyFile, func() {}, nil
		}
		if err := ensureTLSDir(certFile, keyFile); err != nil {
			return "", "", nil, err
		}
		if err := writeSelfSignedCertificate(certFile, keyFile); err != nil {
			return "", "", nil, err
		}
		return certFile, keyFile, func() {}, nil
	}

	tempDir, err := os.MkdirTemp("", "fitcommerce-tls-*")
	if err != nil {
		return "", "", nil, fmt.Errorf("create temp tls dir: %w", err)
	}

	cleanup = func() {
		_ = os.RemoveAll(tempDir)
	}

	resolvedCertFile = filepath.Join(tempDir, "server.crt")
	resolvedKeyFile = filepath.Join(tempDir, "server.key")

	if err := writeSelfSignedCertificate(resolvedCertFile, resolvedKeyFile); err != nil {
		cleanup()
		return "", "", nil, err
	}

	return resolvedCertFile, resolvedKeyFile, cleanup, nil
}

func filesExist(paths ...string) bool {
	for _, path := range paths {
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			return false
		}
	}
	return true
}

func ensureTLSDir(certFile, keyFile string) error {
	certDir := filepath.Dir(certFile)
	keyDir := filepath.Dir(keyFile)
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		return fmt.Errorf("create cert dir: %w", err)
	}
	if err := os.MkdirAll(keyDir, 0o755); err != nil {
		return fmt.Errorf("create key dir: %w", err)
	}
	return nil
}

func writeSelfSignedCertificate(certPath, keyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate rsa key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate certificate serial number: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "fitcommerce.local",
			Organization: []string{"FitCommerce"},
		},
		NotBefore:             time.Now().UTC().Add(-1 * time.Hour),
		NotAfter:              time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "fitcommerce.local", "fitcommerce-api", "backend"},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
		},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		return fmt.Errorf("write certificate file: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return fmt.Errorf("write key file: %w", err)
	}

	return nil
}
