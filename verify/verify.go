package verify

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func Verify(c echo.Context) error {
	fmt.Println("Verify Endpoint Hit")

	// get signature and encoded cert from header
	signature := c.Request().Header.Get("Signature")
	encodedCert := c.Request().Header.Get("Cert")
	if signature == "" || encodedCert == "" {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	// redirect the request to the CA
	// TODO: modify ca url
	caURL := "http://localhost:1111/verify"

	req := c.Request().Clone(context.Background())
	req.URL, _ = url.Parse(caURL)
	req.RequestURI = ""

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusUnauthorized, err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	decodedCert, _ := base64.StdEncoding.DecodeString(encodedCert)
	cert, _ := x509.ParseCertificate(decodedCert)

	// pubKeyAlgo := cert.PublicKeyAlgorithm
	// use ed25519
	pubKey := cert.PublicKey.(*ed25519.PublicKey)
	if ed25519.Verify(*pubKey, decodedCert, []byte(signature)) {
		username := cert.Subject.CommonName
		InitSession(c, username)
		return c.String(http.StatusOK, "OK")
	} else {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
}

func Filter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip filtering for special endpoints
		if c.Path() == "/" && c.Request().Method == "GET" {
			return next(c)
		}

		sess, _ := session.Get("session", c)
		ifValid := sess.Values["valid"]

		if ifValid != "valid" {
			return Verify(c)
		}
		return next(c)
	}
}

func InitSession(c echo.Context, username string) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   120, // 2 minutes for testing
		HttpOnly: true,
	}
	sess.Values["valid"] = "valid"
	sess.Values["username"] = username
	sess.Save(c.Request(), c.Response())
}
