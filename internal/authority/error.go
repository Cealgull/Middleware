package authority

import (
	"net/http"

	"github.com/Cealgull/Middleware/internal/proto"
)

type CertDecodeError struct{}
type CertFormatError struct{}
type CertUnauthorizedError struct{}
type CertInternalError struct{}
type CertMissingError struct{}
type SignatureDecodeError struct{}
type SignatureMissingError struct{}
type SignatureVerificationError struct{}

func (e *CertInternalError) Error() string {
	return "Cert: Internal Server Error."
}

func (e *CertInternalError) Status() int {
	return http.StatusInternalServerError
}

func (e *CertInternalError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1001",
		Message: e.Error(),
	}
}

func (e *CertDecodeError) Error() string {
	return "Cert: Certifiate Decode Error. Please verify your input."
}

func (e *CertDecodeError) Status() int {
	return http.StatusBadRequest
}

func (e *CertDecodeError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1002",
		Message: e.Error(),
	}
}

func (e *CertMissingError) Error() string {
	return "Cert: Cert Missing Error. Please verify your body."
}

func (e *CertMissingError) Status() int {
	return http.StatusBadRequest
}

func (e *CertMissingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1003",
		Message: e.Error(),
	}
}

func (e *CertFormatError) Error() string {
	return "Cert: Certifiate Format Error. Please verify your input."
}

func (e *CertFormatError) Status() int {
	return http.StatusBadRequest
}

func (e *CertFormatError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "C1003",
		Message: e.Error(),
	}
}

func (e *CertUnauthorizedError) Error() string {
	return "Cert: Unauthorized Certificate. Not Signed by Verify."
}

func (e *CertUnauthorizedError) Status() int {
	return http.StatusUnauthorized
}

func (e *CertUnauthorizedError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "A0240",
		Message: e.Error(),
	}
}

func (e *SignatureDecodeError) Error() string {
	return "Signature: Signature Decode Error. Please verify your input."
}

func (e *SignatureDecodeError) Status() int {
	return http.StatusBadRequest
}

func (e *SignatureDecodeError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "S1001",
		Message: e.Error(),
	}
}

func (e *SignatureMissingError) Error() string {
	return "Signature: Signature Missing Error. Please verify your headers."
}

func (e *SignatureMissingError) Status() int {
	return http.StatusBadRequest
}

func (e *SignatureMissingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "S1002",
		Message: e.Error(),
	}
}

func (e *SignatureVerificationError) Error() string {
	return "Signature: Signature Missing Error. Please verify your headers."
}

func (e *SignatureVerificationError) Status() int {
	return http.StatusBadRequest
}

func (e *SignatureVerificationError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "S1002",
		Message: e.Error(),
	}
}
