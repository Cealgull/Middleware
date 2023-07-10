package main

import (
	"Cealgull_middleware/firefly"
	"Cealgull_middleware/ipfs"
	"net/http"
	"net/url"

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
	e = echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())

	ipfs.Init("localhost:6001")

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/uploadFile", ipfs.UploadFile)
	e.POST("/uploadString", ipfs.UploadString)

	e.POST("/register", firefly.Register)

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

	e.Logger.Fatal(e.Start(":1323"))
}
