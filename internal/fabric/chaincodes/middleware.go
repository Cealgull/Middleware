package chaincodes

import (
	"github.com/Cealgull/Middleware/internal/fabric/common"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type ChaincodeInvoke func(contract common.Contract, c echo.Context) error

type ChaincodeEventCallback func(payload []byte) error

type ChaincodeQuery echo.HandlerFunc

type ChaincodeCustom func(contract common.Contract, c echo.Context) error

type ChaincodeMiddleware struct {
	name     string
	net      common.Network
	contract common.Contract

	invokes   map[string]ChaincodeInvoke
	callbacks map[string]ChaincodeEventCallback
	queries   map[string]ChaincodeQuery

	custom map[string]ChaincodeCustom
	logger *zap.Logger
}

type ChaincodeMiddlewareOption func(cc *ChaincodeMiddleware) error

func WithChaincodeHandler(action string, eventName string, invoke ChaincodeInvoke, callback ChaincodeEventCallback) ChaincodeMiddlewareOption {
	return func(cc *ChaincodeMiddleware) error {
		cc.invokes[action] = invoke
		cc.callbacks[eventName] = callback
		return nil
	}
}

func WithChaincodeQuery(token string, query ChaincodeQuery) ChaincodeMiddlewareOption {
	return func(cc *ChaincodeMiddleware) error {
		cc.queries[token] = query
		return nil
	}
}

func WithChaincodeCustom(location string, custom ChaincodeCustom) ChaincodeMiddlewareOption {
	return func(cc *ChaincodeMiddleware) error {
		cc.custom[location] = custom
		return nil
	}
}

func NewChaincodeMiddleware(logger *zap.Logger, net common.Network, contract common.Contract, options ...ChaincodeMiddlewareOption) *ChaincodeMiddleware {
	cc := ChaincodeMiddleware{
		name:      contract.ChaincodeName(),
		net:       net,
		contract:  contract,
		invokes:   make(map[string]ChaincodeInvoke),
		callbacks: make(map[string]ChaincodeEventCallback),
		queries:   make(map[string]ChaincodeQuery),
		custom:    make(map[string]ChaincodeCustom),
		logger:    logger,
	}

	for _, option := range options {
		var _ = option(&cc)
	}
	return &cc
}

func (cc *ChaincodeMiddleware) Register(g *echo.Group, e *echo.Echo) {

	i := g.Group("/invoke")

	for action, invoke := range cc.invokes {
		i.POST("/"+action, func(invoke ChaincodeInvoke) echo.HandlerFunc {
			return func(c echo.Context) error { return invoke(cc.contract, c) }
		}(invoke))
	}

	q := g.Group("/query")

	for action, query := range cc.queries {
		q.POST("/"+action, func(query ChaincodeQuery) echo.HandlerFunc {
			return func(c echo.Context) error { return query(c) }
		}(query))
	}

	for location, custom := range cc.custom {
		e.POST(location, func(custom ChaincodeCustom) echo.HandlerFunc {
			return func(c echo.Context) error { return custom(cc.contract, c) }
		}(custom))
	}

}

func (cc *ChaincodeMiddleware) Listen(ctx context.Context) error {

	ch, err := cc.net.ChaincodeEvents(ctx, cc.contract.ChaincodeName())

	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-ch:
			callback, _ := cc.callbacks[event.EventName]
			cc.logger.Info("Received Ledger Event", zap.String("name", event.EventName))
			if callback != nil {
				go func(data []byte) {
					if err := callback(data); err != nil {
						cc.logger.Error("Error when calling event callback", zap.Error(err))
					}
				}(event.Payload)
			}
		}
	}
}
