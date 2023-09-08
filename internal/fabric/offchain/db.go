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

func WithPostgresGormConfig(config *config.PostgresGormConfig) PostgresOption {
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

func NewOffchainStore(dialector gorm.Dialector, config *config.PostgresGormConfig) (*gorm.DB, error) {

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

	if config.Prometheus.Enabled {
		var _ = db.Use(prometheus.New(
			prometheus.Config{
				DBName:         "cealgull",
				StartServer:    true,
				HTTPServerPort: uint32(config.Prometheus.Port),
			},
		))
	}

	if config.Seed {
		if !db.Migrator().HasTable(&User{}) {
			user := &User{Username: "admin", Wallet: "0x0994edd4e11125c26df5ac32553de95c569934cf32f594f54090b423", Avatar: ""}
			var _ = db.Create(user)
			var _ = db.Create(&Profile{Signature: "this is admin", UserWallet: &user.Wallet})
			var _ = db.Create(&CategoryGroup{Name: "General", Color: "#4F709C"})
			var _ = db.Create(&Category{CategoryGroupName: "General", Name: "General Discussion", Color: "#E5D283"})
			var _ = db.Create(&Category{CategoryGroupName: "General", Name: "General Topic", Color: "#213555"})
			var _ = db.Create(&[]Tag{{Name: "Tag1"}, {Name: "Tag2"}, {Name: "Tag3"}, {Name: "Tag4"}, {Name: "Tag5"}})
		}
	}

	return db, err
}
