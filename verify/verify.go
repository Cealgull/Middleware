package verify

import (
	"Cealgull_middleware/config"

	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var Config config.MiddlewareConfig

func Verify(c echo.Context) error {
	fmt.Println("Verify Endpoint Hit")

	// get signature from header, and cert from body
	encodedSignature := c.Request().Header.Get("Signature")
	decodedSignature, err := base64.StdEncoding.DecodeString(encodedSignature)
	if err != nil {
		fmt.Println("failed to decode signature", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	jsonMap := make(map[string]interface{})
	err = json.NewDecoder(c.Request().Body).Decode(&jsonMap)
	if err != nil {
		return c.String(http.StatusBadRequest, "Parse JSON error")
	}
	reqCert := jsonMap["cert"].(string)
	if reqCert == "" {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	/*
		fmt.Println("Signature:", decodedSignature)
		fmt.Println("Cert:", reqCert)
	*/

	// redirect the request to the CA
	// uncomment the following code if CA is ready
	/*
		caURL := Config.Ca.Url
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
	*/

	block, _ := pem.Decode([]byte(reqCert))
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Println("failed to decode PEM block containing certificate")
		fmt.Printf("block: %v\ntype: %s\n", block, block.Type)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	x509cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Println("failed to parse certificate", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	/*
		fmt.Println("cert:", x509cert)
		fmt.Println("cert.PublicKey:", x509cert.PublicKey)
		fmt.Println("cert.Subject.CommonName:", x509cert.Subject.CommonName)
	*/
	fmt.Println("cert.Subject.CommonName:", x509cert.Subject.CommonName)

	// use ed25519 as the crypto algorithm
	pubKey := x509cert.PublicKey.(ed25519.PublicKey)
	if ed25519.Verify(pubKey, block.Bytes, decodedSignature) {
		userId := x509cert.Subject.CommonName
		InitSession(c, userId)
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

func InitSession(c echo.Context, userId string) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   120, // 2 minutes for testing
		HttpOnly: true,
	}
	sess.Values["valid"] = "valid"
	sess.Values["userId"] = userId
	sess.Save(c.Request(), c.Response())
}

func Login(c echo.Context) error {
	fmt.Println("Login Endpoint Hit")
	// TODO: return userprofile (if not exist, register)
	return nil
}
