package domain

// User ...
type User struct {
	ID             uint64 `json:"id;omitempty"`    // get - owner only
	Email          string `json:"email;omitempty"` // create; get - owner only
	Username       string `json:"username"`        // create; get - public
	Name           string `json:"name"`
	URL            string `json:"url"`
	ProfileImgURL  string `json:"profile_img_url"`
	Location       string `json:"location"`
	Description    string `json:"description"`
	IsMe           bool   `json:"isme"`
	FollowersCount int    `json:"followers_count"`
	FollowingCount int    `json:"following_count"`
	TwitterName    string `json:"twitter_name"`
	FacebookName   string `json:"facebook_name"`
}

// GetURL ...
func (user *User) GetURL() string {
	if len(user.URL) == 0 {
		user.URL = "/@" + user.Username
	}
	return "/@" + user.Username
}

// UserService defines interface that a user-service layer
// can provide as use-cases
type UserService interface {

	// User getter/query interfaces
	GetUserProfile(username string) (User, error)
	// GetRecommendUsers() ([]User, error) // no RecSys yet.

	// User writer interfaces
	GetOrCreateUser(email string, user User) (User, error)
	DeleteUser(userID uint64) error

	// User updater interfaces
	UpdateUsername(userID uint64, user User) (User, error)

	// User update social media link
	// TODO: AddTwitterName(twitterName string)
	// TODO: AddFacebookName(fbName string)

	// User follows other user, the followed.
	FollowUser(userID uint64, followedUsername string) (User, error)
	UnfollowUser(userID uint64, followedUsername string) (User, error)

	// User Image Profile
	// TODO: UploadProfileImage()
}

// UserRepository defines interface that user-data
// persistence layer (db, cache, elastic) can provide
type UserRepository interface {

	// Query single user
	GetByID(userID uint64) (User, error)
	GetByEmail(email string) (User, error)
	GetByUsername(username string) (User, error)

	// Query paginate many users
	// TODO:FetchMany(userFilter User, page int, limit int) ([]User, Metadata, error)

	// Insert single user
	InsertOne(user User) (User, error)

	// Update single user
	UpdateOne(userID uint64, user User) (User, error)

	// Relate user follower-followed relationship
	//TODO: RelateUsers(followedID uint64, followerID uint64) (error)

	// Delete single user
	DeleteOne(userID uint64) error
}
