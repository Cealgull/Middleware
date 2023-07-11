package main

import (
	"Cealgull_middleware/firefly"
	"Cealgull_middleware/ipfs"
	"Cealgull_middleware/verify"

	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/exp/slices"
)

var e *echo.Echo

var specialEndpoints = []string{
	"/",
	"/register",
	"/uploadFile",
	"/uploadString",
}

func main() {
	ipfs.Init("localhost:6001")

	e = echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())

	url1, err := url.Parse("http://localhost:5000")
	if err != nil {
		e.Logger.Fatal(err)
	}

	url2, err := url.Parse("http://localhost:5001")
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Skipper: func(c echo.Context) bool {
			return slices.Contains(specialEndpoints, c.Path())
		},
		Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
			{
				URL: url1,
			},
			{
				URL: url2,
			},
		}),
	}))

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(verify.Filter)

	e.GET("/", func(c echo.Context) error {
		verify.InitSession(c, "test")
		return c.String(http.StatusOK, "You are logged in now")
	})
	e.POST("/", func(c echo.Context) error {
		sess, _ := session.Get("session", c)
		ifValid := sess.Values["valid"]

		if ifValid != "valid" {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/uploadFile", ipfs.UploadFile)
	e.POST("/uploadString", ipfs.UploadString)
	e.POST("/register", firefly.Register)

	e.Logger.Fatal(e.Start(":1323"))
}
