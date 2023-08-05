package chaincodes

import (
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type ChaincodeInvoke func(contract *client.Contract, c echo.Context) error

type ChaincodeEventCallback func(payload []byte) error

type ChaincodeMiddleware struct {
	name      string
	net       *client.Network
	contract  *client.Contract
	cc_invokes   map[string]ChaincodeInvoke
	cc_callbacks map[string]ChaincodeEventCallback
  api_invokes  map[string]echo.HandlerFunc
	logger    *zap.Logger
}

type ChaincodeMiddlewareOption func(cc *ChaincodeMiddleware) error

func WithChaincodeHandler(action string, eventName string, invoke ChaincodeInvoke, callback ChaincodeEventCallback) ChaincodeMiddlewareOption {
	return func(cc *ChaincodeMiddleware) error {
		cc.cc_invokes[action] = invoke
		cc.cc_callbacks[eventName] = callback
		return nil
	}
}

func NewChaincodeMiddleware(logger *zap.Logger, net *client.Network, ccName string, options ...ChaincodeMiddlewareOption) *ChaincodeMiddleware {
	cc := ChaincodeMiddleware{
		name:      ccName,
		net:       net,
		contract:  net.GetContract(ccName),
		cc_invokes:   make(map[string]ChaincodeInvoke),
		cc_callbacks: make(map[string]ChaincodeEventCallback),
		logger:    logger,
	}

	for _, option := range options {
		var _ = option(&cc)
	}
	return &cc
}

func (cc *ChaincodeMiddleware) registerInvoke(action string, invoke ChaincodeInvoke) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := invoke(cc.contract, c)
			return err
		}
}

func (cc *ChaincodeMiddleware) Register(g *echo.Group) {
  
  i := g.Group("/invoke")

	for action, invoke := range cc.cc_invokes {
		i.POST("/"+action, cc.registerInvoke(action, invoke))
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
			callback, _ := cc.cc_callbacks[event.EventName]
			cc.logger.Info("Received Ledger Event", zap.String("name", event.EventName))
			if callback != nil {
				callback(event.Payload)
			}
		}
	}
}
