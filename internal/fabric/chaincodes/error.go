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

type ChaincodeBase64DecodeError struct{}

func (f *ChaincodeBase64DecodeError) Error() string {
	return "Chaincode: Failed to decode base64 data."
}

func (f *ChaincodeBase64DecodeError) Status() int {
	return http.StatusBadRequest
}

func (f *ChaincodeBase64DecodeError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1004",
		Message: f.Error(),
	}
}

type ChaincodeFieldValidationFailure struct {
  field string
}


func (f *ChaincodeFieldValidationFailure) Error() string {
  return "Chaincode: Failed to validate field " + f.field
}

func (f *ChaincodeFieldValidationFailure) Status() int {
  return http.StatusBadRequest
}

func (f *ChaincodeFieldValidationFailure) Message() *proto.ResponseMessage {
  return &proto.ResponseMessage{
    Code:    "C1005",
    Message: f.Error(),
  }
}


var success *proto.Success = &proto.Success{}
var chaincodeInternalError *ChaincodeInternalError = &ChaincodeInternalError{}
var chaincodeDeserializationError *ChaincodeDeserializationError = &ChaincodeDeserializationError{}
var chaincodeBase64DecodeError *ChaincodeBase64DecodeError = &ChaincodeBase64DecodeError{}
