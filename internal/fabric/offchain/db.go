package offchain

import (
	"fmt"
	"github.com/Cealgull/Middleware/internal/config"
	"github.com/Cealgull/Middleware/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabaseStore(config *config.PostgresConfig) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.New(
		postgres.Config{
			DSN: fmt.Sprintf("host=%s port=%d user=%s dbname=%s",
				config.Host,
				config.Port,
				config.User,
				config.Name),
		}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{},
		&models.Profile{},
		&models.Role{},
		&models.Badge{}); err != nil {
		return nil, err
	}

	return db, err
}
