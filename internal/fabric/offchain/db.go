package offchain

import (
	"fmt"

	"github.com/Cealgull/Middleware/internal/config"
	. "github.com/Cealgull/Middleware/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
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

func NewOffchainStore(dialector gorm.Dialector, promethusConfig *config.PrometheusConfig) (*gorm.DB, error) {

	db, err := gorm.Open(dialector, &gorm.Config{
		FullSaveAssociations: true,
		Logger:               logger.Default.LogMode(logger.Warn),
	})

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(

		Role{},
		Badge{},
		Upvote{},
		Downvote{},
		Asset{},

		User{},
		Profile{},
		Topic{},
		Post{},
		Tag{},
		TagRelation{},
		OwnedToken{},
		TradedToken{},

		CategoryGroup{},
		Category{},
		CategoryRelation{},
		RoleRelation{},
		BadgeRelation{},
	); err != nil {
		return nil, err
	}

	if promethusConfig.Enabled {
		var _ = db.Use(prometheus.New(
			prometheus.Config{
				DBName:         "cealgull",
				StartServer:    true,
				HTTPServerPort: uint32(promethusConfig.Port),
			},
		))
	}

	return db, err
}
