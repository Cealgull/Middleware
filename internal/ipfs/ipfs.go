package ipfs

import (
	"io"

	"github.com/Cealgull/Middleware/internal/proto"
	IPFS "github.com/ipfs/go-ipfs-api"
	"go.uber.org/zap"
)

type IPFSManager struct {
	sh     *IPFS.Shell
	logger *zap.Logger
}

type Option func(mgr *IPFSManager) error

func NewIPFSManager(logger *zap.Logger, url string) (*IPFSManager, error) {

	var mgr IPFSManager
	sh := IPFS.NewShell(url)
	logger.Info("Initializing the ipfs shell", zap.String("URL", url))

	_, _, err := sh.Version()

	if err != nil {
		logger.Error("Error happened when initializing IPFS", zap.String("Error", err.Error()))
		return nil, ipfsBackendError
	}

	mgr.sh = sh
	mgr.logger = logger

	return &mgr, nil
}

func (m *IPFSManager) Put(payload io.Reader) (string, proto.MiddlewareError) {
	cid, err := m.sh.Add(payload)
	if err != nil {
		return "", ipfsBackendError
	}
	return cid, nil
}

func (m *IPFSManager) Cat(cid string) ([]byte, proto.MiddlewareError) {
	r, err := m.sh.Cat(cid)
	if err != nil {
		return nil, &IPFSFileNotFoundError{}
	}
	data, _ := io.ReadAll(r)

	return data, nil
}
