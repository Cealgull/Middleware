package chaincodes

import (
	"reflect"

	"github.com/Cealgull/Middleware/internal/proto"
	"gorm.io/gorm"
)

func validate(db *gorm.DB, models interface{}, names []string) proto.MiddlewareError {

	if len(names) == 0 {
		return nil
	}

	if err := db.Where("name IN ?", names).Find(&models).Error; reflect.ValueOf(models).Len() != len(names) || err != nil {
		name := reflect.TypeOf(models).Elem().Name()
		chaincodeFieldValidationError := &ChaincodeFieldValidationError{name}
		return chaincodeFieldValidationError
	}

	return nil
}

func paginate(pageOrdinal int, pageSize int) func(db *gorm.DB) *gorm.DB{
  return func(db *gorm.DB) *gorm.DB {
    return db.Offset((pageOrdinal - 1) * pageSize).Limit(pageSize)
  }
}
