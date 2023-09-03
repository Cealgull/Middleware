package ipfs

import (
	"fmt"
	"io"

	"github.com/Cealgull/Middleware/internal/proto"
	ipfs "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type IPFSStorage interface {
	Version() (string, string, error)
	Add(payload io.Reader, opts ...ipfs.AddOpts) (string, error)
	Cat(cid string) (io.ReadCloser, error)
}

type IPFSManager struct {
	storage IPFSStorage
	logger  *zap.Logger
}

type IPFSManagerOption func(mgr *IPFSManager) error

func WithIPFSStorage(storage IPFSStorage) IPFSManagerOption {
	return func(mgr *IPFSManager) error {
		mgr.storage = storage
		return nil
	}
}

func WithUrl(url string, port int) IPFSManagerOption {
	return func(mgr *IPFSManager) error {
    sh := ipfs.NewShell(fmt.Sprintf("%s:%d", url, port))
		mgr.storage = sh
		return nil
	}
}

func NewIPFSManager(logger *zap.Logger, options ...IPFSManagerOption) (*IPFSManager, error) {

	mgr := IPFSManager{logger: logger}

	for _, option := range options {
		var _ = option(&mgr)
	}

	_, _, err := mgr.storage.Version()
	if err != nil {
		return nil, err
	}

	return &mgr, nil
}

func (m *IPFSManager) Put(payload io.Reader) (string, proto.MiddlewareError) {
	cid, err := m.storage.Add(payload)
	if err != nil {
		return "", ipfsBackendError
	}
	return cid, nil
}

func (m *IPFSManager) Cat(cid string) ([]byte, proto.MiddlewareError) {
	r, err := m.storage.Cat(cid)
	if err != nil {
		return nil, &StorageFileNotFoundError{}
	}
	data, _ := io.ReadAll(r)

	return data, nil
}

func (m *IPFSManager) upload(c echo.Context) error {

	type UploadResponse struct {
		Cid string `json:"cid"`
	}

	payload, err := c.FormFile("payload")

	if err != nil {
		return c.JSON(uploadFileMissingError.Status(),
			uploadFileMissingError.Message())
	}

	file, _ := payload.Open()

	if cid, err := m.Put(file); err != nil {
		return c.JSON(err.Status(), err.Message())
	} else {
		uploadResponse := UploadResponse{Cid: cid}
		return c.JSON(success.Status(), &uploadResponse)
	}

}

func (im *IPFSManager) Register(echo *echo.Echo) error {
	echo.POST("/api/upload", im.upload)
	return nil
}
