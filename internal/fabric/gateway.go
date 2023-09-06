package fabric

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Cealgull/Middleware/internal/config"
	"github.com/Cealgull/Middleware/internal/fabric/chaincodes"
	"github.com/Cealgull/Middleware/internal/fabric/offchain"
	"github.com/Cealgull/Middleware/internal/ipfs"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/gorm"
)

type GatewayMiddleware struct {
	db     *gorm.DB
	cm     map[string]*chaincodes.ChaincodeMiddleware
	logger *zap.Logger
}

func loadCertificate(certPath string) (*x509.Certificate, error) {

	p, err := os.ReadDir(certPath)

	if err != nil {
		return nil, fmt.Errorf("Failed to read certificate path: %w", err)
	}

	b, err := os.ReadFile(path.Join(certPath, p[0].Name()))

	if err != nil {
		return nil, fmt.Errorf("Failed to open certificate file: %w", err)
	}

	return identity.CertificateFromPEM(b)
}

func loadSign(keyPath string) (identity.Sign, error) {

	p, err := os.ReadDir(keyPath)

	if err != nil {
		return nil, fmt.Errorf("Failed to read certificate path: %w", err)
	}

	b, err := os.ReadFile(path.Join(keyPath, p[0].Name()))

	if err != nil {
		return nil, fmt.Errorf("Failed to open certificate file: %w", err)
	}

	pk, err := identity.PrivateKeyFromPEM(b)

	if err != nil {
		return nil, fmt.Errorf("Failed to open load private key: %w", err)
	}

	return identity.NewPrivateKeySign(pk)
}

func initNetwork(config *config.GatewayConfig) (*client.Network, error) {

	tlsPath := path.Join(config.CryptoPath, config.User, "msp/cacerts")
	certPath := path.Join(config.CryptoPath, config.User, "msp/signcerts")
	keyPath := path.Join(config.CryptoPath, config.User, "msp/keystore")

	tlsCert, err := loadCertificate(tlsPath)

	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCert)

	tlsCred := credentials.NewClientTLSFromCert(certPool, config.GatewayPeer)

	connection, err := grpc.Dial(config.PeerEndpoint,
		grpc.WithTransportCredentials(tlsCred))

	if err != nil {
		return nil, err
	}

	cert, err := loadCertificate(certPath)

	if err != nil {
		return nil, err
	}

	sign, err := loadSign(keyPath)

	if err != nil {
		return nil, err
	}

	id, err := identity.NewX509Identity(config.MspID, cert)

	if err != nil {
		return nil, err
	}

	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(connection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)

	if err != nil {
		return nil, err
	}

	return gateway.GetNetwork(config.Channel), nil
}

func NewGatewayMiddleware(logger *zap.Logger, ipfs *ipfs.IPFSManager, config *config.MiddlewareConfig) (*GatewayMiddleware, error) {

	dialector := offchain.NewPostgresDialector(offchain.WithPostgresDSNConfig(&config.Postgres))

	db, err := offchain.NewOffchainStore(dialector)

	if err != nil {
		return nil, err
	}

	network, err := initNetwork(&config.Gateway)

	if err != nil {
		return nil, err
	}

	cm := make(map[string]*chaincodes.ChaincodeMiddleware)

	cm["user"] = chaincodes.NewUserProfileMiddleware(logger, network, db)
	cm["topic"] = chaincodes.NewTopicChaincodeMiddleware(logger, network, ipfs, db)
  cm["post"] = chaincodes.NewPostChaincodeMiddleware(logger, network, ipfs, db)
  cm["tag"] = chaincodes.NewTagChaincodeMiddleware(logger, network, ipfs, db)
  cm["category"] = chaincodes.NewCategoryChaincodeMiddleware(logger, network, ipfs, db)
  cm["categoryGroup"] = chaincodes.NewCategoryGroupChaincodeMiddleware(logger, network, ipfs, db)

	return &GatewayMiddleware{
		db:     db,
		cm:     cm,
		logger: logger,
	}, nil

}

func (g *GatewayMiddleware) Register(e *echo.Echo) error {
	for n, m := range g.cm {
		c := e.Group("/api/" + n)
		m.Register(c, e)
		go func(m *chaincodes.ChaincodeMiddleware) {
			m.Listen(context.Background())
		}(m)
	}

	return nil
}
