package authority

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/jarcoal/httpmock"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func generateRequest(method string, url string, sig string, cert string) echo.Context {

	var b []byte

	if len(cert) != 0 {
		b, _ = json.Marshal(&CACert{Cert: cert})
	} else {
		b = []byte("abcd")
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(b))
  if len(cert) != 0{
	  req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
  }
	req.Header.Add("signature", sig)
	rec := httptest.NewRecorder()

	c := server.NewContext(req, rec)
	c.SetPath(url)
	return c
}

func methodOKhandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func TestValidateSession(t *testing.T) {

	httpmock.RegisterResponder("POST", ca.endpoint, httpmock.NewStringResponder(200, `"OK`))
	defer httpmock.Reset()

	s := session.Middleware(sessions.NewCookieStore([]byte("secret")))
	v := ca.ValidateSession(methodOKhandler)
	v = s(v)

	c := generateRequest("GET", "/", "", "")
	err := v(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	c = generateRequest("POST", "/auth/login", "", "")
	err = v(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	sig, cert := generateCert(t)

	c = generateRequest("POST", "/auth/login", sig, "")
	err = v(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	c = generateRequest("POST", "/auth/login", "", cert)
	err = v(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

  _, cert2 := generateCert(t)
  
	c = generateRequest("POST", "/auth/login", sig, cert2)
	err = v(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Response().Status)

	c = generateRequest("POST", "/auth/login", sig, cert)
	err = v(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, c.Response().Status)

	sess, _ := session.Get("session", c)
	assert.True(t, sess.Values["authorized"].(bool))
}
