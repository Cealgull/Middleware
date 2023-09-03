package authority

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var ca *CertAuthority
var server *echo.Echo

const HOST = "api.cealgull.verify"
const PORT = 80

func generateCert(t *testing.T) (string, string) {
	pub, priv, _ := ed25519.GenerateKey(nil)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Cealgull",
		},
		Issuer: pkix.Name{
			Organization: []string{"cealgull"},
		},
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	assert.NoError(t, err)

	pemcert := pem.EncodeToMemory(&pem.Block{
		Type:  CERT,
		Bytes: cert,
	})

	sig := ed25519.Sign(priv, pemcert)

	sigb64 := base64.StdEncoding.EncodeToString(sig)

	return sigb64, string(pemcert)
}


func TestValidateCertInternalError(t *testing.T) {
	cert, err := ca.validateCert("test", CACert{Cert: "test"})

  var _ = err.Message()
  var _ = err.Status()

	assert.IsType(t, &CertInternalError{}, err)
	assert.Nil(t, cert)
}

func TestValidateCertUnauthorized(t *testing.T) {

	httpmock.RegisterResponder("POST", ca.endpoint,
		httpmock.NewStringResponder(500, "verification error"))

	defer httpmock.Reset()

	cert, err := ca.validateCert("test", CACert{Cert: "test"})

  var _ = err.Message()
  var _ = err.Status()

	assert.IsType(t, &CertUnauthorizedError{}, err)
	assert.Nil(t, cert)

}

func TestValidateCertNoExternal(t *testing.T) {

	httpmock.RegisterResponder("POST", ca.endpoint,
		httpmock.NewStringResponder(200, "OK"))

	defer httpmock.Reset()

	cert, err := ca.validateCert("[]ssqs", CACert{Cert: "abcd"})

  var _ = err.Message()
  var _ = err.Status()

	assert.IsType(t, &SignatureDecodeError{}, err)
	assert.Nil(t, cert)

	sig1, _ := generateCert(t)
	sig2, cert2 := generateCert(t)

	cert, err = ca.validateCert(sig1, CACert{Cert: cert2})

  var _ = err.Message()
  var _ = err.Status()

	assert.IsType(t, &SignatureVerificationError{}, err)
	assert.Nil(t, cert)

	cert, err = ca.validateCert(sig2, CACert{Cert: cert2})

	assert.NoError(t, err)
	assert.NotNil(t, cert)
}

func TestMain(m *testing.M) {

	logger, _ := zap.NewProduction()
	ca = NewCertAuthority(logger, HOST, PORT)
	httpmock.ActivateNonDefault(ca.client.GetClient())
  server = echo.New()
  ca.Register(server)

	defer httpmock.DeactivateAndReset()
	var _ = m.Run()
}
