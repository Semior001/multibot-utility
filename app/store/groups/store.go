package groups

//go:generate mockery -inpkg -name Store -case snake

// Store defines methods to store and fetch user groups
type Store interface {
	PutGroup(chatID string, alias string, users []string) (err error)
	AddUser(chatID string, alias string, user string) (err error)
	GetGroup(chatID string, alias string) (users []string, err error)
	GetGroups(chatID string) (groups map[string][]string, err error)
	DeleteUserFromGroup(chatID string, alias string, user string) (err error)
	DeleteGroup(chatID string, alias string) (err error)
}
