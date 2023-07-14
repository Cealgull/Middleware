package verify

import (
	"Cealgull_middleware/config"
	"Cealgull_middleware/firefly"
	"bytes"
	"io"
	"net/http"

	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type CACert struct {
	Cert string `json:"cert"`
}

var Config config.MiddlewareConfig

func Verify(c echo.Context) error {
	fmt.Println("Verify Endpoint Hit")

	// get signature from header, and cert from body
	encodedSignature := c.Request().Header.Get("Signature")
	decodedSignature, err := base64.StdEncoding.DecodeString(encodedSignature)
	if err != nil {
		return errors.New("failed to decode signature: " + err.Error())
	}

	var reqCert CACert
	if err := c.Bind(&reqCert); err != nil {
		return err
	}

	caURL := Config.Ca.Url
	b, _ := json.Marshal(reqCert)
	caReq, err := http.NewRequest("POST", caURL, bytes.NewBuffer(b))
	if err != nil {
		return errors.New("failed to create request to CA")
	}
	caReq.Header.Set("content-type", "application/json")
	caReq.Header.Set("Signature", encodedSignature)
	client := &http.Client{}
	caRes, err := client.Do(caReq)
	if err != nil {
		return errors.New("send cert to CA failed " + err.Error())
	}
	if caRes.StatusCode != http.StatusOK {
		bytes, _ := io.ReadAll(caRes.Body)
		return errors.New("ca verify failed " + string(bytes))
	}
	fmt.Println("CA verified", caRes.StatusCode, caRes.Body)

	block, _ := pem.Decode([]byte(reqCert.Cert))
	if block == nil || block.Type != "CERTIFICATE" {
		return errors.New("failed to decode PEM block containing certificate")
	}

	x509cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("failed to parse certificate: " + err.Error())
	}
	fmt.Println("cert.Subject.CommonName:", x509cert.Subject.CommonName)

	// use ed25519 as the crypto algorithm
	pubKey := x509cert.PublicKey.(ed25519.PublicKey)
	if ed25519.Verify(pubKey, []byte(reqCert.Cert), decodedSignature) {
		userId := x509cert.Subject.CommonName
		InitSession(c, userId)
		return nil
	} else {
		return errors.New("failed to verify signature")
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
			err := Verify(c)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}
		}
		return next(c)
	}
}

func InitSession(c echo.Context, userId string) {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 60 minutes for testing
		HttpOnly: true,
	}
	sess.Values["valid"] = "valid"
	sess.Values["userId"] = userId
	sess.Save(c.Request(), c.Response())
}

func Login(c echo.Context) error {
	fmt.Println("Login Endpoint Hit")
	sess, _ := session.Get("session", c)
	userId := sess.Values["userId"]

	readUserRes, err := firefly.ReadUserLogin(c, userId.(string))
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer readUserRes.Body.Close()

	if readUserRes.StatusCode == 200 {
		return c.Stream(http.StatusOK, "application/json", readUserRes.Body)
	}

	registerRes, err := firefly.Register(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer registerRes.Body.Close()

	fmt.Println("registerRes.StatusCode:", registerRes.StatusCode)

	if registerRes.StatusCode == http.StatusAccepted {
		count := 0
		for count < 30 {
			count++
			// wait for the user to be registered
			time.Sleep(100 * time.Millisecond)
			readUserRes, err := firefly.ReadUserLogin(c, userId.(string))
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			defer readUserRes.Body.Close()

			if readUserRes.StatusCode == 200 {
				return c.Stream(http.StatusOK, "application/json", readUserRes.Body)
			}
		}
	}
	return c.String(http.StatusInternalServerError, "failed to login")
}
