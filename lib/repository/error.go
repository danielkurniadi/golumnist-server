package repository

import (
	"regexp"

	"github.com/iqdf/golumn-story-service/domain"
	"github.com/jinzhu/gorm"
)

// GormErrConverter ...
type GormErrConverter struct {
	dialect string
}

// NewGormErrCvt ...
func NewGormErrCvt(dialect string) *GormErrConverter {
	return &GormErrConverter{dialect: dialect}
}

func (errCvt *GormErrConverter) checkNoRecordError(dbErr error) bool {
	return (dbErr == gorm.ErrRecordNotFound)
}

func (errCvt *GormErrConverter) checkTransactionError(dbErr error) bool {
	return (dbErr == gorm.ErrInvalidTransaction ||
		dbErr == gorm.ErrCantStartTransaction)
}

func (errCvt *GormErrConverter) checkSQLError(dbErr error) bool {
	return dbErr == gorm.ErrInvalidSQL
}

func (errCvt *GormErrConverter) checkUnaddressedError(dbErr error) bool {
	return dbErr == gorm.ErrUnaddressable
}

// AppError converts Gorm based error to domain.AppError
func (errCvt *GormErrConverter) AppError(dbErr error, message string) error {
	if dbErr == nil {
		return nil
	}

	switch { // switch condition == true
	case errCvt.checkNoRecordError(dbErr):
		return domain.ErrUnknownResource.WithMessage(
			"item not found with specified identifier/field")

	case errCvt.checkTransactionError(dbErr):
		return domain.ErrInternalServer.Wrap(dbErr, message)

	case errCvt.checkSQLError(dbErr):
		return domain.ErrInternalServer.Wrap(dbErr, message)

	case errCvt.checkUnaddressedError(dbErr):
		return domain.ErrInternalServer.Wrap(dbErr,
			"repository: fatal human error: use reference &DBModel{} in gorm: db.CallSomething(model)")

	default:
		return domain.ErrInternalServer.Wrap(dbErr, "repository: fatal uncaught db error")
	}
}

var (
	// RegexpMySQLDuplicate matches error string when
	// MySQL DB throw integrity/duplicate error during insert operation
	RegexpMySQLDuplicate = regexp.MustCompile(`^Error (?P<Code>\d{4}): Duplicate entry '(?P<Field>.+)' for key '(?P<Value>.+)'$`)

	// RegexpMySQLDataLength matches error string when
	// MySQL DB throw data length error during insert operation
	RegexpMySQLDataLength = regexp.MustCompile(`Error (?P<Code>\d{4}): Data too long for column '(?P<Field>)' at row (?P<Row>.+)$`)
)

// MySQLErrConverter ...
type MySQLErrConverter struct {
	GormErrConverter
}

// NewMySQLErrCvt ...
func NewMySQLErrCvt() *MySQLErrConverter {
	return &MySQLErrConverter{
		GormErrConverter: *NewGormErrCvt("mysql"),
	}
}

func (errCvt *MySQLErrConverter) checkDuplicateError(dbErr error) bool {
	return RegexpMySQLDuplicate.Match([]byte(dbErr.Error()))
}

func (errCvt *MySQLErrConverter) checkDataLengthError(dbErr error) bool {
	return RegexpMySQLDataLength.Match([]byte(dbErr.Error()))
}

// AppError ...
func (errCvt *MySQLErrConverter) AppError(dbErr error, message string) error {
	if dbErr == nil {
		return nil
	}

	if appErr := errCvt.GormErrConverter.AppError(dbErr, message); appErr != nil {
		return appErr
	}

	switch { // switch condition == true
	case errCvt.checkDuplicateError(dbErr):
		rexGroup := getParams(*RegexpMySQLDuplicate, dbErr.Error())
		field, _ := rexGroup["Field"]
		return domain.ErrBadParameters.WithMessagef("conflict duplicate %v", field)

	case errCvt.checkDataLengthError(dbErr):
		rexGroup := getParams(*RegexpMySQLDataLength, dbErr.Error())
		field, _ := rexGroup["Field"]
		return domain.ErrBadParameters.WithMessagef("data too long for %v field", field)

	default:
		return domain.ErrInternalServer.Wrap(dbErr, message)
	}
}

func getParams(regEx regexp.Regexp, str string) (paramsMap map[string]string) {
	match := regEx.FindStringSubmatch(str)
	paramsMap = make(map[string]string)

	for i, name := range regEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return
}
