package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"
)

// KeyAlgo selects the key type/size for generated certificates and CSRs.
type KeyAlgo string

const (
	KeyRSA2048   KeyAlgo = "rsa-2048"
	KeyRSA4096   KeyAlgo = "rsa-4096"
	KeyECDSAP256 KeyAlgo = "ecdsa-p256"
	KeyECDSAP384 KeyAlgo = "ecdsa-p384"
)

// CertRequest describes a certificate or CSR to generate.
type CertRequest struct {
	CommonName   string
	Organization string
	Country      string
	DNSNames     []string
	IPs          []string // dotted/textual IPs; invalid entries are ignored
	ValidFor     time.Duration
	Algo         KeyAlgo
}

// GeneratedCert bundles the PEM-encoded certificate + private key and the parsed
// certificate for immediate inspection.
type GeneratedCert struct {
	CertPEM     []byte
	KeyPEM      []byte
	Certificate *x509.Certificate
}

// GenerateSelfSigned creates a self-signed certificate and its private key.
func GenerateSelfSigned(req CertRequest) (*GeneratedCert, error) {
	priv, err := newKey(req.Algo)
	if err != nil {
		return nil, err
	}
	serial, err := randSerial()
	if err != nil {
		return nil, err
	}
	validFor := req.ValidFor
	if validFor <= 0 {
		validFor = 365 * 24 * time.Hour
	}
	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               subject(req),
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(validFor),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	applySANs(tmpl, req)

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, publicKey(priv), priv)
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}
	keyPEM, err := encodeKeyPEM(priv)
	if err != nil {
		return nil, err
	}
	return &GeneratedCert{
		CertPEM:     pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		KeyPEM:      keyPEM,
		Certificate: cert,
	}, nil
}

// GenerateCSR creates a PKCS#10 certificate signing request and its private key,
// both PEM-encoded.
func GenerateCSR(req CertRequest) (csrPEM, keyPEM []byte, err error) {
	priv, err := newKey(req.Algo)
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.CertificateRequest{Subject: subject(req)}
	if len(req.DNSNames) > 0 {
		tmpl.DNSNames = req.DNSNames
	}
	for _, ip := range req.IPs {
		if parsed := net.ParseIP(ip); parsed != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, parsed)
		}
	}
	der, err := x509.CreateCertificateRequest(rand.Reader, tmpl, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("create CSR: %w", err)
	}
	keyPEM, err = encodeKeyPEM(priv)
	if err != nil {
		return nil, nil, err
	}
	csrPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der})
	return csrPEM, keyPEM, nil
}

// ParseCertificatePEM decodes the first CERTIFICATE block from PEM input.
func ParseCertificatePEM(pemBytes []byte) (*x509.Certificate, error) {
	for {
		block, rest := pem.Decode(pemBytes)
		if block == nil {
			return nil, errors.New("no CERTIFICATE block found in PEM input")
		}
		if block.Type == "CERTIFICATE" {
			return x509.ParseCertificate(block.Bytes)
		}
		pemBytes = rest
	}
}

// ExpiryLevel classifies how close a certificate is to expiry, driving the
// colour-coded expiry watchdog in the UI.
type ExpiryLevel string

const (
	ExpiryOK       ExpiryLevel = "ok"       // > 30 days
	ExpiryWarn     ExpiryLevel = "warning"  // <= 30 days
	ExpiryCritical ExpiryLevel = "critical" // <= 7 days
	ExpiryExpired  ExpiryLevel = "expired"  // past NotAfter
)

// ExpiryStatus reports days remaining and a severity level.
type ExpiryStatus struct {
	NotAfter time.Time
	DaysLeft int
	Level    ExpiryLevel
}

// CheckExpiry evaluates cert against the reference time now.
func CheckExpiry(cert *x509.Certificate, now time.Time) ExpiryStatus {
	days := int(cert.NotAfter.Sub(now).Hours() / 24)
	level := ExpiryOK
	switch {
	case !now.Before(cert.NotAfter):
		level = ExpiryExpired
	case days <= 7:
		level = ExpiryCritical
	case days <= 30:
		level = ExpiryWarn
	}
	return ExpiryStatus{NotAfter: cert.NotAfter, DaysLeft: days, Level: level}
}

// --- helpers ---

func subject(req CertRequest) pkix.Name {
	n := pkix.Name{CommonName: req.CommonName}
	if req.Organization != "" {
		n.Organization = []string{req.Organization}
	}
	if req.Country != "" {
		n.Country = []string{req.Country}
	}
	return n
}

func applySANs(tmpl *x509.Certificate, req CertRequest) {
	tmpl.DNSNames = req.DNSNames
	for _, ip := range req.IPs {
		if parsed := net.ParseIP(ip); parsed != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, parsed)
		}
	}
}

func newKey(algo KeyAlgo) (any, error) {
	switch algo {
	case KeyRSA4096:
		return rsa.GenerateKey(rand.Reader, 4096)
	case KeyECDSAP256:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case KeyECDSAP384:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case KeyRSA2048, "":
		return rsa.GenerateKey(rand.Reader, 2048)
	default:
		return nil, fmt.Errorf("unsupported key algorithm %q", algo)
	}
}

func publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func encodeKeyPEM(priv any) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

func randSerial() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, limit)
}
