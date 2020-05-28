package mysql

import (
	// import built-in libraries
	"reflect"
	"regexp"
	"strings"
	"time"

	// import third-party libraries
	"github.com/jinzhu/gorm"

	// import our local package
	"github.com/iqdf/golumn-story-service/domain"
)

// DefaultLimit ...
const DefaultLimit uint = 20

// UserDB ...
type UserDB struct {
	ID            uint64 `gorm:"primary_key"`
	Email         string `gorm:"Type:VARCHAR(15);UNIQUE_INDEX;NOT_NULL"`
	Username      string `gorm:"Type:VARCHAR(15);UNIQUE_INDEX;NOT NULL"`
	Name          string `gorm:"Type:VARCHAR(15);INDEX;NOT NULL"`
	ProfileImgURL string `gorm:"Type:VARCHAR(20);DEFAULT:'/icon/defaultauthor'"`
	Location      string `gorm:"Type:VARCHAR(10);DEFAULT:'Worldwide'"`
	Description   string `gorm:"Type:VARCHAR(128);DEFAULT:'The author tend to keep air of mystery of him/herself'"`
	Followers     int
	Following     int
	TwitterName   string `gorm:"Type:VARCHAR(20)"`
	FacebookName  string `gorm:"Type:VARCHAR(20)"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewUserDBWriter ...
func NewUserDBWriter(user domain.User) UserDB {
	return UserDB{
		Email:         user.Email,
		Username:      user.Username,
		Name:          user.Name,
		ProfileImgURL: user.ProfileImgURL,
		Location:      user.Location,
		Description:   user.Description,
		Followers:     0,
		Following:     0,
		TwitterName:   "",
		FacebookName:  "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
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
		column := ToSnakeCase(val.Type().Field(i).Name)
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
		ID:            userDB.ID,
		Email:         userDB.Email,
		Username:      userDB.Username,
		Name:          userDB.Name,
		ProfileImgURL: userDB.ProfileImgURL,
		Location:      userDB.Location,
		Description:   userDB.Description,
		Followers:     userDB.Followers,
		Following:     userDB.Following,
		TwitterName:   userDB.TwitterName,
		FacebookName:  userDB.FacebookName,
	}
}

// UIntRandomizer ...
// TODO: put under common
type UIntRandomizer interface {
	Uint32() uint32
	Uint64() uint64
}

// UserMySQLRepository ...
type UserMySQLRepository struct {
	DB   *gorm.DB
	Rand UIntRandomizer
}

// NewUserMySQLRepository ...
func NewUserMySQLRepository(db *gorm.DB, rand UIntRandomizer) *UserMySQLRepository {
	return &UserMySQLRepository{
		DB:   db,
		Rand: rand,
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
	db = db.Where("id = ?", userID).First(&userDB)
	return userDB.User(), db.Error
}

// GetByEmail ...
func (userRepo *UserMySQLRepository) GetByEmail(email string) (domain.User, error) {
	var (
		userDB = new(UserDB)
		db     = userRepo.DB
	)
	// SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` LIMIT 1
	db = db.Where("email = ?", email).First(&userDB)
	return userDB.User(), db.Error
}

// GetByUsername ...
func (userRepo *UserMySQLRepository) GetByUsername(username string) (domain.User, error) {
	var (
		userDB = new(UserDB)
		db     = userRepo.DB
	)
	// SELECT * FROM `users` WHERE (username = ?) ORDER BY `users`.`id` LIMIT 1
	db = db.Where("username = ?", username).First(&userDB)
	return userDB.User(), db.Error
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
		return domain.User{}, err
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
	db = db.Model(&userDB).Updates(userDB) // update attributes of a user row
	if err := db.Error; err != nil || db.RowsAffected == 0 {
		return domain.User{}, err
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
		return err
	}
	return nil
}
