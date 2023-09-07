package offchain

import (
	"errors"
	"fmt"

	"github.com/Cealgull/Middleware/internal/config"
	. "github.com/Cealgull/Middleware/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func NewOffchainStore(dialector gorm.Dialector, seed bool) (*gorm.DB, error) {

	db, err := gorm.Open(dialector, &gorm.Config{
		FullSaveAssociations: true,
		Logger:               logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	if !db.Migrator().HasTable(&User{}) {
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
	}

	if seed {
		if db.Migrator().HasTable(&Tag{}) {
			wallet := "wallet"
			if err := db.First(&User{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
				var _ = db.Create(&User{Username: "admin", Wallet: wallet}).Error
				var _ = db.Create(&Profile{UserWallet: &wallet}).Error
				var _ = db.Create(&[]Tag{
					{CreatorWallet: wallet, Name: "tag1"},
					{CreatorWallet: wallet, Name: "tag2"}}).Error
				var _ = db.Create(&CategoryGroup{Name: "categoryGroup1", Color: "#D17898",
					Categories: []*Category{
						{Name: "category1", Color: "#A2C0BF"},
						{Name: "category2", Color: "#FA827D"},
					}}).Error
			}
		}
	}

	return db, err
}
