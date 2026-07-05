package security

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

func TestGenerateSelfSignedRSA(t *testing.T) {
	gc, err := GenerateSelfSigned(CertRequest{
		CommonName:   "conduit.local",
		Organization: "Conduit",
		DNSNames:     []string{"conduit.local", "*.conduit.local"},
		IPs:          []string{"127.0.0.1", "not-an-ip"},
		ValidFor:     48 * time.Hour,
		Algo:         KeyRSA2048,
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if gc.Certificate.Subject.CommonName != "conduit.local" {
		t.Fatalf("CN: got %q", gc.Certificate.Subject.CommonName)
	}
	if len(gc.Certificate.DNSNames) != 2 {
		t.Fatalf("DNSNames: got %v", gc.Certificate.DNSNames)
	}
	if len(gc.Certificate.IPAddresses) != 1 { // invalid IP dropped
		t.Fatalf("IPs: got %v", gc.Certificate.IPAddresses)
	}

	// PEM must round-trip back to an equivalent certificate.
	parsed, err := ParseCertificatePEM(gc.CertPEM)
	if err != nil {
		t.Fatalf("parse PEM: %v", err)
	}
	if !parsed.NotAfter.Equal(gc.Certificate.NotAfter) {
		t.Fatal("NotAfter mismatch after PEM round-trip")
	}
}

func TestGenerateSelfSignedECDSA(t *testing.T) {
	gc, err := GenerateSelfSigned(CertRequest{CommonName: "ec.local", Algo: KeyECDSAP256})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if gc.Certificate.PublicKeyAlgorithm != x509.ECDSA {
		t.Fatalf("expected ECDSA key, got %v", gc.Certificate.PublicKeyAlgorithm)
	}
}

func TestGenerateCSR(t *testing.T) {
	csrPEM, keyPEM, err := GenerateCSR(CertRequest{
		CommonName: "csr.local", Organization: "Conduit",
		DNSNames: []string{"csr.local"}, Algo: KeyECDSAP256,
	})
	if err != nil {
		t.Fatalf("csr: %v", err)
	}
	if len(keyPEM) == 0 {
		t.Fatal("empty key PEM")
	}
	block := firstPEM(t, csrPEM, "CERTIFICATE REQUEST")
	csr, err := x509.ParseCertificateRequest(block)
	if err != nil {
		t.Fatalf("parse CSR: %v", err)
	}
	if err := csr.CheckSignature(); err != nil {
		t.Fatalf("CSR signature invalid: %v", err)
	}
	if csr.Subject.CommonName != "csr.local" {
		t.Fatalf("CSR CN: got %q", csr.Subject.CommonName)
	}
}

func TestCheckExpiry(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		after time.Time
		want  ExpiryLevel
	}{
		{"healthy", base.Add(60 * 24 * time.Hour), ExpiryOK},
		{"warn", base.Add(20 * 24 * time.Hour), ExpiryWarn},
		{"critical", base.Add(3 * 24 * time.Hour), ExpiryCritical},
		{"expired", base.Add(-1 * time.Hour), ExpiryExpired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cert := &x509.Certificate{NotAfter: tc.after}
			got := CheckExpiry(cert, base)
			if got.Level != tc.want {
				t.Fatalf("level: got %s want %s (daysLeft=%d)", got.Level, tc.want, got.DaysLeft)
			}
		})
	}
}

func firstPEM(t *testing.T, data []byte, typ string) []byte {
	t.Helper()
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			t.Fatalf("no %s block found", typ)
		}
		if block.Type == typ {
			return block.Bytes
		}
		data = rest
	}
}
