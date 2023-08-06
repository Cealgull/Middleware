package chaincodes

import (
	"github.com/Cealgull/Middleware/internal/ipfs"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func invokeCreatePost(logger *zap.Logger, ipfs *ipfs.IPFSManager, db *gorm.DB) ChaincodeInvoke {
  return func(contract *client.Contract, c echo.Context) error {
    return nil
  }
}
