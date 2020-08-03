package repeater

import "github.com/google/uuid"

//go:generate mockery -inpkg -name Store -case snake

// Store defines methods to put and load regexp rules and chats
type Store interface {
	Add(r Rule) error                 // adds rule to storage
	Find(req FindReq) ([]Rule, error) // returns list of rules (or its ids) by specified request
	Delete(id string) error           // removes rule from storage by its ID
}

// Service wraps Store with additional logic needed for all Store implementations
type Service struct {
	Store
}

// Add checks whether id is defined and if not defined, assigns it and gives a control flow to the
// store's implementation of Store.Add method
func (s *Service) Add(r Rule) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return s.Store.Add(r)
}

// FindReq describes optional parameters to get rules action.
// At least one parameter required
type FindReq struct {
	Author string // get rules by author
	Src    string // get rules by source
}

// Rule describes a basic repeating rule
type Rule struct {
	ID string // id of the given rule

	Src  string // source chat of repeating messages
	Re   string // regular expression to filter repeating messages
	Dest string // destination of repeating messages

	Author string // author of this rule
}
