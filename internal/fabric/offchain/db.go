package offchain

import (
	"errors"
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

	if err := db.Model(&User{}).First(&User{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		wallet := "0x12345678"
		var _ = db.Create(&User{Username: "admin", Avatar: "", Wallet: wallet})
		var _ = db.Create(&Profile{Signature: "this is a signature", UserWallet: &wallet})
		var _ = db.Create(&CategoryGroup{Name: "General", Color: "#E25E3E"})
		var _ = db.Create(&Category{CategoryGroupName: "General", Name: "General Discussion", Color: "#E25E3E"})
		var _ = db.Create(&Category{CategoryGroupName: "General", Name: "General Topic", Color: "#C63D2F"})
		var _ = db.Create(&[]*Tag{{CreatorWallet: "0x12345678", Name: "tag1"},
			{CreatorWallet: "0x12345678", Name: "tag2"},
			{CreatorWallet: "0x12345678", Name: "tag3"}})
		var _ = db.Create(&Topic{Title: "this is a test topic",
			Hash:             "hash1",
			Content:          "Genshin Impact is a good game",
			CreatorWallet:    "0x12345678",
			CategoryAssigned: &CategoryRelation{CategoryName: "General Topic"}})
    var _ = db.Create(&Post{Hash: "post1Hash", CreatorWallet: "0x12345678", BelongToHash: "hash1", Content: "this is a test post"})
		id := uint(1)
    var _ = db.Create(&Post{Hash: "post2Hash", CreatorWallet: "0x12345678", BelongToHash: "hash1", Content: "this is a test post", ReplyToID: &id})
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

	return db, err
}
