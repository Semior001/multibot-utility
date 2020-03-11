package bot

import (
	"github.com/Semior001/multibotUtility/app/store/groups"
)

// GroupBot gathers usernames into one mention, like @admins
type GroupBot struct {
	Store *groups.Store
}

func NewGroupBot(store *groups.Store) *GroupBot {
	return &GroupBot{
		Store: store,
	}
}

func (g *GroupBot) OnMessage(msg Message) *Response {
	panic("implement me")
}

func (g *GroupBot) ReactsOn() []string {
	panic("implement me")
}

func (g *GroupBot) Help() string {
	return `/add_group @<group alias> @<user1>, @<user2>, ... - add user
/delete_user_from_group @<group alias> @<user> - removes user from the group
/detete_group @<group alias> - removes group
/list_groups - shows the list of existing groups
/add_user_to_group @<group alias> @<user> - adds user to the specified group
@<group alias> - triggers bot to send message with all participants of the group`
}
