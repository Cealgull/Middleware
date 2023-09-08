package offchain

import (
	"testing"

	"github.com/Cealgull/Middleware/internal/config"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
)

func TestMain(m *testing.M) {
	var _ = m.Run()
}

func TestNewOffchainStore(t *testing.T) {

	t.Run("testing with mock connection with automigrate failure", func(t *testing.T) {

		mockConn, _, _ := sqlmock.New()

		dialector := NewPostgresDialector(WithPostgresConn(mockConn))
		store, err := NewOffchainStore(dialector, &config.PostgresGormConfig{Prometheus: config.PrometheusConfig{Enabled: true, Port: 9090}})

		assert.Error(t, err)
		assert.Nil(t, store)

	})

	t.Run("testing with mock connection with automigrate success", func(t *testing.T) {

		dialector := sqlite.Open("file::memory:")

		store, err := NewOffchainStore(dialector, &config.PostgresGormConfig{Prometheus: config.PrometheusConfig{Enabled: true, Port: 9090}})

		assert.NoError(t, err)
		assert.NotNil(t, store)
	})

	t.Run("testing with dsn config", func(t *testing.T) {
		c := config.PostgresGormConfig{}
		dialector := NewPostgresDialector(WithPostgresGormConfig(&c))
		_, err := NewOffchainStore(dialector, &config.PostgresGormConfig{Prometheus: config.PrometheusConfig{Enabled: true, Port: 9090}})
		assert.Error(t, err)
	})
}
