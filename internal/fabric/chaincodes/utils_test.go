package chaincodes

import (
	"testing"

	. "github.com/Cealgull/Middleware/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestValidate(t *testing.T) {

	user := &User{
		Username: "Admin",
		Wallet:   "0x123456789",
		Avatar:   "null",
	}

	tags := []*Tag{
		{Name: "Genshin Impact", CreatorWallet: user.Wallet, Description: "Genshin Impact"},
		{Name: "Honkai Impact", CreatorWallet: user.Wallet, Description: "Honkai Impact"},
	}

	db := newSqliteDB()

	assert.NoError(t, db.Create(&user).Error)
	assert.NoError(t, db.Create(&tags).Error)

	t.Run("Validating with Empty Tags", func(t *testing.T) {
		err := validate(db, []Tag{}, []string{})
		assert.NoError(t, err)
	})

	t.Run("Validating with Success", func(t *testing.T) {
		assert.NoError(t, validate(db, []Tag{}, []string{"Genshin Impact", "Honkai Impact"}))
	})

	t.Run("Validating with Failure", func(t *testing.T) {
		err := validate(db, []Tag{}, []string{"Genshin Impact 2", "Genshin Impact 3"})
		assert.Error(t, err)
		logger.Info("Error:", zap.Error(err))
	})

}
