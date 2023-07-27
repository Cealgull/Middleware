package authority

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"

	"github.com/Cealgull/Middleware/internal/proto"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CACert struct {
	Cert string `json:"cert"`
}

type CertAuthority struct {
	client   *resty.Client
	logger   *zap.Logger
	endpoint string
}

const (
	CERTIFICATE = "CERTIFICATE"
)

func NewCertAuthority(logger *zap.Logger, endpoint string) *CertAuthority {
	ca := &CertAuthority{
		client:   resty.New(),
		logger:   logger,
		endpoint: endpoint,
	}
	return ca
}

func (ca *CertAuthority) validateCert(sigb64 string, reqcert CACert) (*x509.Certificate, proto.MiddlewareError) {

	resp, err := ca.client.R().SetBody(reqcert).Post(ca.endpoint)

	if err != nil {
		return nil, &CertInternalError{}
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, &CertUnauthorizedError{}
	}

	sig, err := base64.StdEncoding.DecodeString(sigb64)

	if err != nil {
		return nil, &SignatureDecodeError{}
	}

	b, _ := pem.Decode([]byte(reqcert.Cert))

	if b == nil || b.Type != "CERTIFICATE" {
		return nil, &CertDecodeError{}
	}

	x509cert, err := x509.ParseCertificate(b.Bytes)

	if err != nil {
		return nil, &CertDecodeError{}
	}

	ca.logger.Info("Requesting identity", zap.String("Common Name", x509cert.Subject.CommonName))

	// use ed25519 as the crypto algorithm
	pubKey := x509cert.PublicKey.(ed25519.PublicKey)

	if ed25519.Verify(pubKey, []byte(reqcert.Cert), sig) {
		return x509cert, nil
	} else {
		return nil, &SignatureVerificationError{}
	}
}

func (ca *CertAuthority) Register(e *echo.Echo) error {
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(ca.ValidateSession)
	return nil
}
