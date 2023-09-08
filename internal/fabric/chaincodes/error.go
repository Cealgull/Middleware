package chaincodes

import (
	"net/http"

	"github.com/Cealgull/Middleware/internal/proto"
)

type ChaincodeInvokeFailureError struct {
	contract string
}

func (f *ChaincodeInvokeFailureError) Error() string {
	return "Chaincode: Failed to invoke chaincode " + f.contract
}

func (f *ChaincodeInvokeFailureError) Status() int {
	return http.StatusInternalServerError
}

func (f *ChaincodeInvokeFailureError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1001",
		Message: f.Error(),
	}
}

type ChaincodeInternalError struct{}

func (f *ChaincodeInternalError) Error() string {
	return "Chaincode: Internal Service error."
}

func (f *ChaincodeInternalError) Status() int {
	return http.StatusInternalServerError
}

func (f *ChaincodeInternalError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1002",
		Message: f.Error(),
	}
}

type ChaincodeDeserializationError struct{}

func (f *ChaincodeDeserializationError) Error() string {
	return "Chaincode: Failed to deserialize data."
}

func (f *ChaincodeDeserializationError) Status() int {
	return http.StatusBadRequest
}

func (f *ChaincodeDeserializationError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1003",
		Message: f.Error(),
	}
}

type ChaincodeFieldValidationError struct {
	field string
}

func (f *ChaincodeFieldValidationError) Error() string {
	return "Chaincode: Failed to validate field " + f.field
}

func (f *ChaincodeFieldValidationError) Status() int {
	return http.StatusBadRequest
}

func (f *ChaincodeFieldValidationError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1005",
		Message: f.Error(),
	}
}

type ChaincodeQueryParameterError struct{}

func (f *ChaincodeQueryParameterError) Error() string {
	return "Chaincode: Wrong PageOrdinal Or PageSize."
}

func (f *ChaincodeQueryParameterError) Status() int {
	return http.StatusBadRequest
}

func (f *ChaincodeQueryParameterError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1006",
		Message: f.Error(),
	}
}

type ChaincodeDuplicatedError struct {
	field string
}

func (f *ChaincodeDuplicatedError) Error() string {
	return "Chaincode: Duplicated " + f.field
}

func (f *ChaincodeDuplicatedError) Status() int {
	return http.StatusBadRequest
}

func (f *ChaincodeDuplicatedError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1007",
		Message: f.Error(),
	}
}

type ChaincodeNotFoundError struct {
	field string
}

func (f *ChaincodeNotFoundError) Error() string {
	return "Chaincode: Not Found " + f.field
}

func (f *ChaincodeNotFoundError) Status() int {
	return http.StatusBadRequest
}

func (f *ChaincodeNotFoundError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1008",
		Message: f.Error(),
	}
}

var success *proto.Success = &proto.Success{}
var chaincodeInternalError *ChaincodeInternalError = &ChaincodeInternalError{}
var chaincodeDeserializationError *ChaincodeDeserializationError = &ChaincodeDeserializationError{}
var chaincodeQueryParameterError *ChaincodeQueryParameterError = &ChaincodeQueryParameterError{}
