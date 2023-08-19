package rest

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type RestServer struct {
	addr      string
	endpoints []RestEndpoint
	echo      *echo.Echo
}

type Option func(r *RestServer) error

func WithLogger(logger *zap.Logger) Option {
	return func(r *RestServer) error {
		r.echo.Use(middleware.RequestLoggerWithConfig(
			middleware.RequestLoggerConfig{
				LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
					logger.Info("request",
						zap.String("URI", v.URI),
						zap.String("URIPath", v.URIPath),
						zap.Duration("Latency", v.Latency),
						zap.String("Host", v.Host),
						zap.String("Protocol", v.Protocol),
						zap.String("ContentLength", v.ContentLength),
						zap.Reflect("Headers", v.Headers),
					)
					return nil
				},
			},
		))
		return nil
	}
}

func WithEndpoint(endpoint RestEndpoint) Option {
	return func(r *RestServer) error {
		r.endpoints = append(r.endpoints, endpoint)
		return nil
	}
}

func NewRestServer(host string, port int, options ...Option) (*RestServer, error) {

	var rest RestServer

	rest.addr = fmt.Sprintf("%s:%d", host, port)
	rest.echo = echo.New()
	rest.echo.Use(middleware.Logger())
	rest.echo.HideBanner = true

	for _, option := range options {
		var _ = option(&rest)
	}

	for _, endpoint := range rest.endpoints {
		var _ = endpoint.Register(rest.echo)
	}

	return &rest, nil
}

func (r *RestServer) Start() {
	r.echo.Logger.Fatal(r.echo.Start(r.addr))
}
