package mocks

import (
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/stretchr/testify/mock"
)

//Custom type alias is not supported by package loading.
//This Mock should be implemented manually.
//See https://github.com/vektra/mockery/issues/331

type MockContract struct {
	mock.Mock
}

func (m *MockContract) Submit(transactionName string, options ...client.ProposalOption) ([]byte, error) {
	args := m.Called(transactionName, options)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockContract) SubmitAsync(transactionName string, options ...client.ProposalOption) ([]byte, *client.Commit, error){
  args := m.Called(transactionName, options)
  return args.Get(0).([]byte), args.Get(1).(*client.Commit), args.Error(2)
}

func (m *MockContract) ChaincodeName() string {
	args := m.Called()
	return args.String(0)
}

func NewMockContract() *MockContract {
	return &MockContract{}
}
