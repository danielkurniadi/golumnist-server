package mysql

import (
	// import built-in libraries
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	// import third-party libraries
	"github.com/jinzhu/gorm"

	// import our local packages
	"github.com/iqdf/golumn-story-service/domain"
	repocommon "github.com/iqdf/golumn-story-service/lib/repository"
)

// DefaultLimit ...
const DefaultLimit uint = 20

// UserDB ...
type UserDB struct {
	ID             uint64    `gorm:"PRIMARY_KEY"`
	Email          string    `gorm:"Type:VARCHAR(40);UNIQUE_INDEX;NOT NULL"`
	Username       string    `gorm:"Type:VARCHAR(40);UNIQUE_INDEX;NOT NULL"`
	Name           string    `gorm:"Type:VARCHAR(40);INDEX;NOT NULL"`
	ProfileImgURL  string    `gorm:"Type:VARCHAR(128);DEFAULT:'/icon/defaultpic'"`
	Location       string    `gorm:"Type:VARCHAR(40);DEFAULT:'Worldwide'"`
	Description    string    `gorm:"Type:VARCHAR(256);DEFAULT:'The author tend to keep air of mystery of him/herself'"`
	Followers      []*UserDB `gorm:"MANY2MANY:followership;JOINTABLE_FOREIGNKEY:follower_id;ASSOCIATION_JOINTABLE_FOREIGNKEY:followed_id"`
	FollowersCount int
	FollowingCount int
	TwitterName    string `gorm:"Type:VARCHAR(20)"`
	FacebookName   string `gorm:"Type:VARCHAR(20)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewUserDBWriter ...
func NewUserDBWriter(user domain.User) UserDB {
	// TODO: check NOT NULL string ''
	// TODO: check minimum length contraint
	// email: RFC 3696 - Section 3
	// username: 5 - 40
	// name: 2 - 40
	return UserDB{
		Email:          user.Email,
		Username:       user.Username,
		Name:           user.Name,
		ProfileImgURL:  user.ProfileImgURL,
		Location:       user.Location,
		Description:    user.Description,
		FollowersCount: 0,
		FollowingCount: 0,
		TwitterName:    "",
		FacebookName:   "",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// NewUserDBUpdater ...
func NewUserDBUpdater(userID uint64, user domain.User) UserDB {
	return UserDB{
		Name:          user.Name,
		ProfileImgURL: user.ProfileImgURL,
		Location:      user.Location,
		Description:   user.Description,
		UpdatedAt:     time.Now(),
	}
}

// ToSnakeCase ....
func ToSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// UserColumns return list of column names
func UserColumns() []string {
	user := UserDB{}
	val := reflect.Indirect(reflect.ValueOf(user))

	columns := make([]string, 0, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)

		if strings.Contains(string(field.Tag), "MANY2MANY") {
			// skip many2many fields as it will not
			// appear as table column
			continue
		}
		column := ToSnakeCase(field.Name)
		columns = append(columns, column)
	}
	return columns
}

// TableName ...
func (userDB *UserDB) TableName() string {
	return "users"
}

// User ...
func (userDB *UserDB) User() domain.User {
	return domain.User{
		ID:             userDB.ID,
		Email:          userDB.Email,
		Username:       userDB.Username,
		Name:           userDB.Name,
		ProfileImgURL:  userDB.ProfileImgURL,
		Location:       userDB.Location,
		Description:    userDB.Description,
		FollowersCount: userDB.FollowersCount,
		FollowingCount: userDB.FollowingCount,
		TwitterName:    userDB.TwitterName,
		FacebookName:   userDB.FacebookName,
	}
}

// UIntRandomizer ...
type UIntRandomizer interface {
	Uint32() uint32
	Uint64() uint64
}

// DBErrorConverter ...
type DBErrorConverter interface {
	AppError(error, string) error
}

// UserMySQLRepository ...
type UserMySQLRepository struct {
	DB     *gorm.DB
	Rand   UIntRandomizer
	ErrCvt DBErrorConverter
}

// NewUserMySQLRepository ...
func NewUserMySQLRepository(db *gorm.DB, rand UIntRandomizer) *UserMySQLRepository {
	return &UserMySQLRepository{
		DB:     db,
		Rand:   rand,
		ErrCvt: repocommon.NewMySQLErrCvt(),
	}
}

func (userRepo *UserMySQLRepository) generateID() uint64 {
	return userRepo.Rand.Uint64()
}

// GetByID ...
func (userRepo *UserMySQLRepository) GetByID(userID uint64) (domain.User, error) {
	var (
		userDB = new(UserDB)
		db     = userRepo.DB
	)
	// SELECT * FROM `users` WHERE (id = ?) ORDER BY `users`.`id` LIMIT 1
	err := db.Where("id = ?", userID).First(&userDB).Error
	appErr := userRepo.ErrCvt.AppError(err, "userrepo: find user by id fail")

	return userDB.User(), appErr
}

// GetByEmail ...
func (userRepo *UserMySQLRepository) GetByEmail(email string) (domain.User, error) {
	var (
		userDB = UserDB{Email: email}
		db     = userRepo.DB
	)
	// SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` LIMIT 1
	err := db.Where("email = ?", email).First(&userDB).Error
	fmt.Println("GetByEmail err", err, "Gorm cannot call Error()")
	appErr := userRepo.ErrCvt.AppError(err, "userrepo: find user by email fail")
	fmt.Println("GetByEmail AppErr", appErr)
	return userDB.User(), appErr
}

// GetByUsername ...
func (userRepo *UserMySQLRepository) GetByUsername(username string) (domain.User, error) {
	var (
		userDB = new(UserDB)
		db     = userRepo.DB
	)
	// SELECT * FROM `users` WHERE (username = ?) ORDER BY `users`.`id` LIMIT 1
	err := db.Where("username = ?", username).First(&userDB).Error
	appErr := userRepo.ErrCvt.AppError(err, "userrepo: find user by username fail")

	return userDB.User(), appErr
}

// InsertOne ...
func (userRepo *UserMySQLRepository) InsertOne(user domain.User) (domain.User, error) {
	var (
		userDB = NewUserDBWriter(user)
		db     = userRepo.DB
	)
	userDB.ID = userRepo.generateID()

	// INSERT INTO `users` (...) VALUES (...)
	db = db.Create(&userDB)
	if err := db.Error; err != nil || db.RowsAffected == 0 {
		appErr := userRepo.ErrCvt.AppError(err, "userrepo: insert one user fail")
		return domain.User{}, appErr
	}
	return userDB.User(), nil
}

// UpdateOne ...
func (userRepo *UserMySQLRepository) UpdateOne(userID uint64, user domain.User) (domain.User, error) {
	var (
		userDB = NewUserDBUpdater(userID, user)
		db     = userRepo.DB
	)

	// UPDATE `users` SET location = (location), description = (description)
	// WHERE id = (userID)
	db = db.Set("gorm:association_save_reference", false)
	db = db.Model(&UserDB{ID: userID}).Updates(userDB) // update attributes of a user row
	if err := db.Error; err != nil || db.RowsAffected == 0 {
		appErr := userRepo.ErrCvt.AppError(err, "userrepo: update one user fail")
		return domain.User{}, appErr
	}

	return userDB.User(), nil
}

// DeleteOne ...
func (userRepo *UserMySQLRepository) DeleteOne(userID uint64) error {
	var (
		userDB = &UserDB{ID: userID}
		db     = userRepo.DB
	)

	// DELETE FROM `users` WHERE id = ?
	db = db.Delete(&userDB)
	if err := db.Error; err != nil || db.RowsAffected == 0 { // do you need to check?
		appErr := userRepo.ErrCvt.AppError(err, "userrepo: delete one user fail")
		return appErr
	}
	return nil
}
