package chaincodes

import (
	"testing"

	"github.com/Cealgull/Middleware/internal/config"
	"github.com/Cealgull/Middleware/internal/fabric/offchain"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var server *echo.Echo
var logger *zap.Logger
var sqliteDB *gorm.DB

func newMockSignedContext(c echo.Context) echo.Context {

	c.Set("_session_store", sessions.NewCookieStore([]byte("secret")))
	s, _ := session.Get("session", c)
	s.Values["wallet"] = "0x123456789"

	return c
}

func newSqliteDB() *gorm.DB {

	dialector := sqlite.Open("file::memory:")

	db, _ := offchain.NewOffchainStore(dialector, &config.PostgresGormConfig{Prometheus: config.PrometheusConfig{Enabled: true, Port: 9090}})

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		panic(err)
	}

	return db
}

func newSqlMockDB() (*gorm.DB, sqlmock.Sqlmock) {
	conn, mock, _ := sqlmock.New()
	dialector := mysql.New(mysql.Config{Conn: conn, SkipInitializeWithVersion: true})
	db, _ := gorm.Open(dialector)
	return db, mock
}

func TestMain(m *testing.M) {
	sqliteDB = newSqliteDB()
	server = echo.New()
	logger, _ = zap.NewProduction()
	var _ = m.Run()
}
