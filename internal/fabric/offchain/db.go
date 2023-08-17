package offchain

import (
	"fmt"
	"github.com/Cealgull/Middleware/internal/config"
	. "github.com/Cealgull/Middleware/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresOption func(config *postgres.Config) error

func WithPostgresDSNConfig(config *config.PostgresDSNConfig) PostgresOption {
	return func(c *postgres.Config) error {
		c.DSN = fmt.Sprintf("host=%s port=%d user=%s dbname=%s", config.Host, config.Port, config.User, config.Name)
		return nil
	}
}

func WithPostgresConn(conn gorm.ConnPool) PostgresOption {
	return func(c *postgres.Config) error {
		c.Conn = conn
		return nil
	}
}

func NewPostgresDialector(options ...PostgresOption) gorm.Dialector {
	dialector := postgres.Config{}
	for _, option := range options {
		var _ = option(&dialector)
	}
	return postgres.New(dialector)
}

func NewOffchainStore(dialector gorm.Dialector) (*gorm.DB, error) {

	db, err := gorm.Open(dialector, &gorm.Config{
		FullSaveAssociations: true,
	})

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		Role{},
		Badge{},
		Profile{},
		Topic{},
		Asset{},
		TopicTag{},
		Post{},
	); err != nil {
		return nil, err
	}

	return db, err
}
