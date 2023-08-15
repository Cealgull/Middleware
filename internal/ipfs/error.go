package ipfs

import (
	"net/http"

	"github.com/Cealgull/Middleware/internal/proto"
)

type StorageBackendError struct{}

func (e *StorageBackendError) Status() int {
	return http.StatusInternalServerError
}

func (e *StorageBackendError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "I0001",
		Message: e.Error(),
	}
}

func (e *StorageBackendError) Error() string {
	return "IPFS: IPFS Storage Backend Error or Timeout."
}

type UploadFileMissingError struct{}

func (e *UploadFileMissingError) Status() int {
	return http.StatusBadRequest
}

func (e *UploadFileMissingError) Error() string {
	return "Bad Request: Missing file payload when uploading."
}

func (e *UploadFileMissingError) Message() *proto.ResponseMessage {
	return &proto.ResponseMessage{
		Code:    "B0002",
		Message: e.Error(),
	}
}

type StorageFileNotFoundError struct{}

func (e *StorageFileNotFoundError) Status() int {
  return http.StatusNotFound
}

func (e *StorageFileNotFoundError) Error() string {
  return "Not Found: File not found."
}

func (e *StorageFileNotFoundError) Message() *proto.ResponseMessage {
  return &proto.ResponseMessage{
    Code:    "B0003",
    Message: e.Error(),
  }
}


var success *proto.Success = &proto.Success{}
var uploadFileMissingError *UploadFileMissingError = &UploadFileMissingError{}
var ipfsBackendError *StorageBackendError = &StorageBackendError{}
