package firefly

import (
	"net/http"

	"github.com/Cealgull/Middleware/internal/proto"
)

var success *proto.Success = &proto.Success{}

type LoginFailureError struct{}
type Base64DecodeError struct{}
type JSONBindingError struct{}
type FireflyInternalError struct{}

func (e *JSONBindingError) Error() string {
	return "Request: Bad Request. Cannot bind JSON Request Body."
}

func (e *JSONBindingError) Status() int {
	return http.StatusBadRequest
}

func (e *JSONBindingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "B0001",
		Message: e.Error(),
	}
}

func (e *LoginFailureError) Error() string {
	return "Login failed. Please retry."
}

func (e *LoginFailureError) Status() int {
	return http.StatusUnauthorized
}

func (e *LoginFailureError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "A1001",
		Message: e.Error(),
	}
}

func (e *Base64DecodeError) Error() string {
	return "Request: Bad Request. Cannot decode Base64 Request Body."
}

func (e *Base64DecodeError) Status() int {
	return http.StatusBadRequest
}

func (e *Base64DecodeError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "B0002",
		Message: e.Error(),
	}
}

func (e *FireflyInternalError) Error() string {
	return "Firefly: Internal Server Error."
}

func (e *FireflyInternalError) Status() int {
	return http.StatusInternalServerError
}

func (e *FireflyInternalError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C0001",
		Message: e.Error(),
	}
}

var jsonBindingError *JSONBindingError = &JSONBindingError{}
var fireflyInternalError *FireflyInternalError = &FireflyInternalError{}
