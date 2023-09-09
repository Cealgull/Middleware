package authority

import (
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var signatureMissingError *SignatureMissingError = &SignatureMissingError{}
var certMissingError *CertMissingError = &CertMissingError{}

func (ca *CertAuthority) signSession(c echo.Context, wallet string) error {
	s, _ := session.Get("session", c)
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 60 minutes for testing
		HttpOnly: true,
	}
	s.Values["authorized"] = true
	s.Values["wallet"] = wallet
	return s.Save(c.Request(), c.Response())
}

func (ca *CertAuthority) ValidateSession(next echo.HandlerFunc) echo.HandlerFunc {

	return func(c echo.Context) error {
		// Skip filtering for special endpoints
		if c.Path() == "/" && c.Request().Method == "GET" {
			return next(c)
		}

		s, _ := session.Get("session", c)

		if v, ok := s.Values["authorized"].(bool); c.Request().URL.RequestURI() == "/auth/login" || !ok || !v {

			var reqcert CACert

			if c.Bind(&reqcert) != nil {
				return c.JSON(certMissingError.Status(), certMissingError.Message())
			}

			signature := c.Request().Header.Get("signature")

			if signature == "" {
				return c.JSON(signatureMissingError.Status(), signatureMissingError.Message())
			}

			cert, err := ca.validateCert(signature, reqcert)

			if err != nil {
				return c.JSON(err.Status(), err.Message())
			}

			var _ = ca.signSession(c, cert.Subject.CommonName)

		}

		return next(c)
	}
}
