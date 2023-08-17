package offchain

import (
	"testing"

	"github.com/Cealgull/Middleware/internal/config"
  "gorm.io/driver/sqlite"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	var _ = m.Run()
}

func TestNewOffchainStore(t *testing.T) {

	t.Run("testing with mock connection with automigrate failure", func(t *testing.T) {

		mockConn, _, _ := sqlmock.New()

		dialector := NewPostgresDialector(WithPostgresConn(mockConn))
		store, err := NewOffchainStore(dialector)

		assert.Error(t, err)
		assert.Nil(t, store)

	})

	t.Run("testing with mock connection with automigrate success", func(t *testing.T) {

    
    dialector := sqlite.Open("file::memory:?cache=shared")

		_, err := NewOffchainStore(dialector)

		store, err := NewOffchainStore(dialector)

		assert.NoError(t, err)
		assert.NotNil(t, store)
	})

	t.Run("testing with dsn config", func(t *testing.T) {
		config := config.PostgresDSNConfig{}
		dialector := NewPostgresDialector(WithPostgresDSNConfig(&config))
		var _, err = NewOffchainStore(dialector)
		assert.Error(t, err)
	})
}
