package chaincodes

import (
	"reflect"

	"github.com/Cealgull/Middleware/internal/proto"
	"gorm.io/gorm"
)

func validate(db *gorm.DB, models interface{}, ids []uint) proto.MiddlewareError {

	if len(ids) == 0 {
		return nil
	}

	if err := db.Find(&models, ids).Error; reflect.ValueOf(models).Len() != len(ids) || err != nil {
		name := reflect.TypeOf(models).Elem().Name()
		chaincodeFieldValidationError := &ChaincodeFieldValidationError{name}
		return chaincodeFieldValidationError
	}

	return nil
}
