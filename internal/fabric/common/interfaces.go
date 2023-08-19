package common
import (
	"context"
	client "github.com/hyperledger/fabric-gateway/pkg/client"
)

type Network interface {
	GetContract(chaincodeName string) *client.Contract
	ChaincodeEvents(ctx context.Context, chaincodeName string, options ...client.ChaincodeEventsOption) (<-chan *client.ChaincodeEvent, error)
}

type Contract interface {
	Submit(transactionName string, options ...client.ProposalOption) ([]byte, error)
	ChaincodeName() string
}
