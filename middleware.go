package main

import (
	"Cealgull_middleware/ipfs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())

	ipfs.Init("localhost:6001")

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/upload", ipfs.Upload)

	e.Logger.Fatal(e.Start(":1323"))
}
