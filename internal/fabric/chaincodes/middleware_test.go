package chaincodes

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Cealgull/Middleware/internal/fabric/common/mocks"
	"github.com/Cealgull/Middleware/internal/models"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func NewMockChaincodeMiddleware(t *testing.T) (*ChaincodeMiddleware, *mocks.MockNetwork) {
	network := mocks.NewMockNetwork(t)
	var _ = network.EXPECT().GetContract("userprofile").Return(&client.Contract{}).Once()
	return NewUserProfileMiddleware(logger, network, newSqliteDB()), network
}

func TestChaincodeMiddlewareRegister(t *testing.T) {
	var m, _ = NewMockChaincodeMiddleware(t)
	m.Register(server.Group("/api"), server)
}

func TestChaincodeMiddlewareListen(t *testing.T) {

	t.Run("Listen and close successfully", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())
		m, network := NewMockChaincodeMiddleware(t)

		cc := make(chan *client.ChaincodeEvent, 5)
		defer cancel()

		go func() {
			network.EXPECT().ChaincodeEvents(ctx, "").Return(cc, nil)
			m.Listen(ctx)
		}()

		time.Sleep(10)

	})

	t.Run("Listen and dealing with successfully", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())
		m, network := NewMockChaincodeMiddleware(t)

		cc := make(chan *client.ChaincodeEvent, 5)
		defer cancel()

		cc <- &client.ChaincodeEvent{
			EventName: "CreateUser",
			Payload:   []byte("abcd"),
		}

		b, _ := json.Marshal(&models.ProfileBlock{
			Username: "Alice",
		})

		cc <- &client.ChaincodeEvent{
			EventName: "CreateUser",
			Payload:   b,
		}

		go func() {
			network.EXPECT().ChaincodeEvents(ctx, "").Return(cc, nil)
			m.Listen(ctx)
		}()

		time.Sleep(1 * time.Second)

	})

	t.Run("Listen with error", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())
		m, network := NewMockChaincodeMiddleware(t)

		defer cancel()

		go func() {
			network.EXPECT().ChaincodeEvents(ctx, "").Return(nil, errors.New("hello world"))
			m.Listen(ctx)
		}()

		time.Sleep(1 * time.Second)

	})

	// network.EXPECT().ChaincodeEvents(ctx, "userprofile").Return()

}
