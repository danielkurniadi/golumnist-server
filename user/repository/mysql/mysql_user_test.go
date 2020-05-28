package mysql

import (
	// import built-in libraries
	"database/sql/driver"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	// import third-party libraries
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-test/deep"
	"github.com/iqdf/golumn-story-service/domain"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AnyTimeArg matches sql Args of type time.Time
// without caring the time value
type AnyTimeArg struct{}

func (a AnyTimeArg) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

// UserIDMocker mocks userID generation in UserRepo such that it
// set the UserDB.ID = 0 which then force DB to autoincrement PK starting
// from PK = 1 onwards
type UserIDMocker struct{}

func NewIDMocker() *UserIDMocker { return &UserIDMocker{} }

func (mockID *UserIDMocker) Uint32() uint32 { return 0 }

func (mockID *UserIDMocker) Uint64() uint64 { return 0 }

type TestSuite struct {
	suite.Suite
	DB         *gorm.DB
	Mock       sqlmock.Sqlmock
	Repository *UserMySQLRepository
}

func (tsuite *TestSuite) SetupSuite() {
	db, mock, err := sqlmock.New()
	require.NoError(tsuite.T(), err)

	tsuite.DB, err = gorm.Open("mysql", db)
	require.NoError(tsuite.T(), err)

	tsuite.DB.SetLogger(log.New(os.Stdout, "\r\n", 0))
	tsuite.DB.LogMode(true)

	tsuite.Mock = mock

	userIDMocker := NewIDMocker()
	tsuite.Repository = NewUserMySQLRepository(tsuite.DB, userIDMocker)
}

func (tsuite *TestSuite) AfterTest(_, _ string) {
	require.NoError(tsuite.T(), tsuite.Mock.ExpectationsWereMet())
	tsuite.DB.Close()
}

func TestInit(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// Warning! Columns order is important!
// MUST Modify this func if UserDB model changes!
// See var UserColumns []string above
func userToRows(user domain.User) []driver.Value {
	return []driver.Value{
		user.ID, user.Email, user.Username,
		user.Name, user.ProfileImgURL,
		user.Location, user.Description,
		user.Followers, user.Following,
		user.TwitterName, user.FacebookName,
		time.Now(), time.Now(),
	}
}

// Warning! Columns order is important! 
// MUST Modify this func if UserDB model changes!
// See NewUserDBWriter() for columns that will be inserted.
func userToInsertArgs(user domain.User) []driver.Value {
	return []driver.Value{
		user.Email, user.Username,
		user.Name, user.ProfileImgURL,
		user.Location, user.Description,
		user.Followers, user.Following,
		user.TwitterName, user.FacebookName,
		AnyTimeArg{}, AnyTimeArg{},
	}
}

var mockUser = domain.User{
	ID:            1,
	Email:         "UserZero-email@example.com",
	Username:      "UserZero",
	Name:          "Name-UserZero",
	ProfileImgURL: "google.profile.com/userzero",
	Location:      "Singapore, Jurong",
	Description:   "AboutMe...",
}

func (tsuite *TestSuite) TestShouldGetByID() {
	rows := sqlmock.NewRows(UserColumns()).
		AddRow(userToRows(mockUser)...)

	queryStr := regexp.QuoteMeta("SELECT * FROM `users` WHERE (id = ?) ORDER BY `users`.`id` ASC LIMIT 1")

	// register sequence of expected operations
	// and defined returned rows to be mocked
	tsuite.Mock.ExpectQuery(queryStr).
		WithArgs(mockUser.ID).
		WillReturnRows(rows)

	// run gorm tx - get user by id
	getUser, err := tsuite.Repository.GetByID(mockUser.ID)
	tsuite.Require().NoError(err)
	tsuite.T().Log("\nDebug Error Log:", err, "\n")
	tsuite.Require().Nil(deep.Equal(getUser, mockUser))
}

func (tsuite *TestSuite) TestShouldGetByEmail() {
	rows := sqlmock.NewRows(UserColumns()).
		AddRow(userToRows(mockUser)...)

	queryStr := regexp.QuoteMeta("SELECT * FROM `users` WHERE (email = ?) ORDER BY `users`.`id` ASC LIMIT 1")

	// register sequence of expected operations
	// and defined returned rows to be mocked
	tsuite.Mock.ExpectQuery(queryStr).
		WithArgs(mockUser.Email).
		WillReturnRows(rows)

	// run gorm tx - get user by id
	getUser, err := tsuite.Repository.GetByEmail(mockUser.Email)
	tsuite.Require().NoError(err)
	tsuite.T().Log("\nDebug Error Log:", err, "\n")
	tsuite.Require().Nil(deep.Equal(getUser, mockUser))
}

func (tsuite *TestSuite) TestShouldGetByUsername() {
	rows := sqlmock.NewRows(UserColumns()).
		AddRow(userToRows(mockUser)...)

	queryStr := regexp.QuoteMeta("SELECT * FROM `users` WHERE (username = ?) ORDER BY `users`.`id` ASC LIMIT 1")

	// register sequence of expected operations
	// and defined returned rows to be mocked
	tsuite.Mock.ExpectQuery(queryStr).
		WithArgs(mockUser.Username).
		WillReturnRows(rows)

	// run gorm tx - get user by id
	getUser, err := tsuite.Repository.GetByUsername(mockUser.Username)
	tsuite.Require().NoError(err)
	tsuite.T().Log("\nDebug Error Log:", err, "\n")
	tsuite.Require().Nil(deep.Equal(getUser, mockUser))
}

func (tsuite *TestSuite) TestShouldInsertOne() {
	insertResult := sqlmock.NewResult(1, 1)
	execStr := regexp.QuoteMeta(
		"INSERT INTO `users` " +
			"(`email`,`username`,`name`,`profile_img_url`,`location`,`description`," +
			"`followers`,`following`,`twitter_name`,`facebook_name`," +
			"`created_at`,`updated_at`) " +
			"VALUES (?,?,?,?,?,?,?,?,?,?,?,?)")

	// register expected tx operations
	// and define mocked db response
	tsuite.Mock.ExpectBegin()
	tsuite.Mock.ExpectExec(execStr).
		WithArgs(userToInsertArgs(mockUser)...).
		WillReturnResult(insertResult)
	tsuite.Mock.ExpectCommit()

	// run gorm tx: insert mock user
	_, err := tsuite.Repository.InsertOne(mockUser)
	tsuite.T().Log("\nDebug Error Log:", err, "\n")
	tsuite.Require().NoError(err)
}

func (tsuite *TestSuite) TestShouldUpdateOne() {
	mockUser := domain.User{
		ID:            1,
		Name:          "Hoi Hoi",
		ProfileImgURL: "https://golumn.com/users/7383/avatar92.jpg",
		Location:      "China, Beijing",
		Description:   "My name is Chalier, I'm test User...",
	}

	updateResult := sqlmock.NewResult(1, 1)

	// order of fields are lexicographically sorted by gorm
	// see NewUserDBUpdater for update-only fields
	execStr := regexp.QuoteMeta("UPDATE `users` SET " +
		"`description` = ?, `location` = ?, " + 
		"`name` = ?, `profile_img_url` = ?")

	args := []driver.Value{
		mockUser.Description, mockUser.Location, 
		mockUser.Name, mockUser.ProfileImgURL,
		AnyTimeArg{},
	}

	// register expected tx operation
	// and defined mocked db response
	tsuite.Mock.ExpectBegin()
	tsuite.Mock.ExpectExec(execStr).
		WithArgs(args...).
		WillReturnResult(updateResult)
	tsuite.Mock.ExpectCommit()

	// run gorm tx: insert mock user
	_, err := tsuite.Repository.UpdateOne(mockUser.ID, mockUser)
	tsuite.T().Log("\nDebug Error Log:", err, "\n")
	tsuite.Require().NoError(err)
}

func (tsuite *TestSuite) TestShouldDeleteOne() {
	var userID uint64 = 1
	deleteResult := sqlmock.NewResult(1, 1)
	execStr := regexp.QuoteMeta("DELETE FROM `users` WHERE `users`.`id` = ?")

	// register expected tx operation
	// and defined mocked db response
	tsuite.Mock.ExpectBegin()
	tsuite.Mock.ExpectExec(execStr).
		WithArgs(userID).
		WillReturnResult(deleteResult)
	tsuite.Mock.ExpectCommit()

	// run gorm tx: insert mock user
	err := tsuite.Repository.DeleteOne(mockUser.ID)
	tsuite.T().Log("\nDebug Error Log:", err, "\n")
	tsuite.Require().NoError(err)
}
