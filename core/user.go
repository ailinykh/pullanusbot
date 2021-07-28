package core

// User ...
type User struct {
	ID           int
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}

func (u *User) DisplayName() string {
	if len(u.Username) == 0 {
		return u.FirstName + " " + u.LastName
	}
	return u.Username
}
