package common

import (
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type Contract interface {
	Submit(transactionName string, options ...client.ProposalOption) ([]byte, error)
	ChaincodeName() string
}

