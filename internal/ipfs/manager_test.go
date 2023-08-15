package ipfs

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var server *echo.Echo

func TestMain(m *testing.M) {
	server = echo.New()
	var _ = m.Run()
}

func newMockIPFSManager(t *testing.T) (*MockIPFSStorage, *IPFSManager) {

	logger, _ := zap.NewProduction()

	s := NewMockIPFSStorage(t)
	s.EXPECT().Version().Return("0.22.0", "ipfs", nil).Once()
	mgr, err := NewIPFSManager(logger, WithIPFSStorage(s))
	assert.NotNil(t, mgr)
	assert.NoError(t, err)

	return s, mgr
}

func TestNewIPFSManager(t *testing.T) {
	var err error

	logger, _ := zap.NewProduction()
	testurl := "localhost:5001"
	mgr, err := NewIPFSManager(logger, WithUrl(testurl))
	assert.Nil(t, mgr)
	assert.Error(t, err)

	var _, _ = newMockIPFSManager(t)
}

func TestIPFSPut(t *testing.T) {

	t.Run("Normal", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		payload := strings.NewReader("test")
		s.EXPECT().Add(payload).Return("QmZv", nil).Once()
		cid, err := mgr.Put(payload)
		assert.Equal(t, "QmZv", cid)
		assert.NoError(t, err)
	})

	t.Run("Backend Error", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		payload := strings.NewReader("test")
		s.EXPECT().Add(payload).Return("", errors.New("hello world")).Once()
		cid, err := mgr.Put(payload)
		assert.Equal(t, "", cid)
		assert.IsType(t, ipfsBackendError, err)
    var _ = err.Status()
    var _ = err.Message()
	})

}

func TestIPFSCat(t *testing.T) {

	t.Run("OK", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		cid := "QmZv"
		s.EXPECT().Cat(cid).Return(io.NopCloser(strings.NewReader("hello world")), nil).Once()
		payload, err := mgr.Cat(cid)
		assert.NotNil(t, payload)
		assert.NoError(t, err)
	})

	t.Run("Storage File Not Found", func(t *testing.T) {
		s, mgr := newMockIPFSManager(t)
		cid := "QmZv"
		s.EXPECT().Cat(cid).Return(nil, errors.New("hello world")).Once()
		payload, err := mgr.Cat(cid)
		assert.Nil(t, payload)
		assert.IsType(t, &StorageFileNotFoundError{}, err)
    var _ = err.Status()
    var _ = err.Message()
	})

}

func TestUpload(t *testing.T) {

	t.Run("Content-type missing", func(t *testing.T) {

		_, mgr := newMockIPFSManager(t)

		req := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		err := mgr.upload(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	})

	t.Run("Payload multipartfile missing error", func(t *testing.T) {

		_, mgr := newMockIPFSManager(t)
		req := httptest.NewRequest(http.MethodPost, "/api/upload", nil)
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		err := mgr.upload(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, c.Response().Status)
	})

	t.Run("Payload multipart ipfs error", func(t *testing.T) {

		s, mgr := newMockIPFSManager(t)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("payload", "test.dat")
		var _, _ = part.Write([]byte("hello world"))

		writer.Close()
		s.Mock.On("Add", mock.Anything).Return("", errors.New("helloworld")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		var _ = mgr.upload(c)
		assert.Equal(t, http.StatusInternalServerError, c.Response().Status)
	})

	t.Run("Payload multipartfile success", func(t *testing.T) {

		s, mgr := newMockIPFSManager(t)
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("payload", "test.dat")
		var _, _ = part.Write([]byte("hello world"))
		writer.Close()
		s.Mock.On("Add", mock.Anything).Return("1234", nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := server.NewContext(req, rec)
		var _ = mgr.upload(c)
		assert.Equal(t, http.StatusOK, c.Response().Status)
	})
}

func TestRegister(t *testing.T) {
	_, mgr := newMockIPFSManager(t)
	var _ = mgr.Register(server)
}
