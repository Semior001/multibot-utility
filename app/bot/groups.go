package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Semior001/multibotUtility/app/store/groups"
)

const regexpAlias = "@[a-zA-Z0-9_]+"

// GroupBot gathers usernames into one mention, like @admins
type GroupBot struct {
	Store              groups.Store
	RespondAllCommands bool
}

// NewGroupBot initializes an instance of GroupBot
func NewGroupBot(store groups.Store, respondAllCmds bool) *GroupBot {
	log.Print("[INFO] GroupBot instantiated")
	return &GroupBot{
		Store:              store,
		RespondAllCommands: respondAllCmds,
	}
}

// OnMessage receives any commands, that are listed in help and group aliases
func (g *GroupBot) OnMessage(msg Message) *Response {
	tokens := strings.Split(msg.Text, " ")

	// command may be in format /cmd@bot
	cmd := strings.Split(tokens[0], "@")[0]
	args := tokens[1:]

	switch cmd {
	case "/add_group":
		if !msg.From.IsAdmin {
			return g.prepareIllegalAccessMessage()
		}
		return g.addGroup(msg, args)
	case "/delete_user_from_group":
		if !msg.From.IsAdmin {
			return g.prepareIllegalAccessMessage()
		}
		return g.deleteUserFromGroup(msg, args)
	case "/delete_group":
		if !msg.From.IsAdmin {
			return g.prepareIllegalAccessMessage()
		}
		return g.deleteGroup(msg, args)
	case "/list_groups":
		return g.listGroups(msg, args)
	case "/add_user_to_group":
		if !msg.From.IsAdmin {
			return g.prepareIllegalAccessMessage()
		}
		return g.addUserToGroup(msg, args)
	}
	return g.handleTrigger(msg)
}

// handleTrigger checks the text for existence of group alias and,
// if present, sends members of it to chat
func (g *GroupBot) handleTrigger(msg Message) *Response {
	// taking all occurrences of aliases, e.g. @admins or @semior001
	seeker, err := regexp.Compile(regexpAlias)
	if err != nil {
		log.Printf("[WARN] error while looking for alias trigger: %+v", err)
		return nil
	}
	byteOccurs := seeker.FindAll([]byte(msg.Text), -1)

	// converting bytes to slice of string aliases
	var aliases []string
	for _, bytes := range byteOccurs {
		aliases = append(aliases, string(bytes))
	}

	// look for aliases in the database
	users, err := g.Store.FindAliases(msg.ChatID, unique(aliases))
	if err != nil {
		log.Printf("[WARN] error while looking for alias trigger %+v", err)
		return nil
	}

	// if no group aliases was found - do nothing
	if len(users) < 1 {
		return nil
	}

	// composing users into one ping message
	return &Response{Text: strings.Join(users, " ")}
}

// addUserToGroup handles /add_user_to_group command and returns corresponding response
// about success or failure executing command
//
// requires exactly two arguments - group alias and username of user, to be added
func (g *GroupBot) addUserToGroup(msg Message, args []string) *Response {
	if len(args) != 2 {
		if g.RespondAllCommands {
			return &Response{
				Reply: true,
				Text:  "Command requires exactly two arguments - group alias and username",
			}
		}
		return nil
	}

	groupAlias := args[0]
	user := args[1]

	err := g.Store.AddUser(msg.ChatID, groupAlias, user)
	if err != nil {
		log.Printf("[WARN] error while adding user to the group %s:%s: %+v", msg.ChatID, groupAlias, err)
		if g.RespondAllCommands {
			return &Response{
				Reply: true,
				Text:  "Internal error",
			}
		}
		return nil
	}

	return &Response{
		Reply: true,
		Text:  fmt.Sprintf("User %s has been successfully added to the group %s", user, groupAlias),
	}
}

// listGroups handles /list_groups command and returns list of existing
// group in this chat
//
// does not require any arguments
func (g *GroupBot) listGroups(msg Message, _ []string) *Response {
	groupList, err := g.Store.GetGroups(msg.ChatID)
	if err != nil {
		log.Printf("[WARN] error while listing groups of chat %s: %+v", msg.ChatID, err)
		if g.RespondAllCommands {
			return &Response{
				Reply: true,
				Text:  "Internal error",
			}
		}
		return nil
	}

	// if no groups are registered in the store - send corresponding response
	if len(groupList) == 0 {
		return &Response{Reply: true, Text: "There's no groups in this chat yet"}
	}

	var groupStrings []string

	// preparing output text in format
	// @group: @user1, @user2, ...
	for alias, users := range groupList {
		groupStrings = append(groupStrings, fmt.Sprintf("%s : %s", alias, strings.Join(users, ", ")))
	}

	return &Response{Reply: true, Text: strings.Join(groupStrings, "\n")}
}

// deleteGroup handles /delete_group command and returns corresponding response
// about success or failure executing command
//
// requires exactly one argument - group alias
func (g *GroupBot) deleteGroup(msg Message, args []string) *Response {
	// command requires exactly one argument - group alias
	if len(args) != 1 {
		if g.RespondAllCommands {
			return &Response{Reply: true, Text: "Command requires exactly one argument - group alias"}
		}
		return nil
	}

	groupAlias := args[0]
	err := g.Store.DeleteGroup(msg.ChatID, groupAlias)
	if err != nil {
		log.Printf("[WARN] error while deleting group %s:%s: %+v", msg.ChatID, groupAlias, err)
		if g.RespondAllCommands {
			return &Response{
				Reply: true,
				Text:  "Internal error",
			}
		}
		return nil
	}
	return &Response{Reply: true, Text: "Group @admins has been successfully deleted"}
}

// deleteUserFromGroup handles /delete_user_from_group command and returns corresponding response
// about success or failure executing command
//
// requires exactly two arguments - group alias and username of user, to be deleted
func (g *GroupBot) deleteUserFromGroup(msg Message, args []string) *Response {
	// command requires exactly one group alias and exactly one username
	if len(args) != 2 {
		if g.RespondAllCommands {
			return &Response{Reply: true, Text: "Command requires exactly two arguments - group alias and username"}
		}
		return nil
	}

	groupAlias := args[0]
	user := args[1]

	err := g.Store.DeleteUserFromGroup(msg.ChatID, groupAlias, user)
	if err != nil {
		log.Printf("[WARN] error while deleting user from group %s:%s: %+v", msg.ChatID, groupAlias, err)
		if g.RespondAllCommands {
			return &Response{
				Reply: true,
				Text:  "Internal error",
			}
		}
		return nil
	}

	return &Response{Reply: true, Text: fmt.Sprintf("User %s has been successfully deleted from group %s", user, groupAlias)}
}

// addGroup handles /add_group command and returns corresponding response
// about success or failure executing command
//
// requires at least two arguments - group alias and usernames
func (g *GroupBot) addGroup(msg Message, args []string) *Response {
	// command requires group alias and at least one username
	if len(args) < 2 {
		if g.RespondAllCommands {
			return &Response{Reply: true, Text: "Not enough parameters to add group"}
		}
		return nil
	}

	groupAlias := args[0]
	users := args[1:]

	err := g.Store.PutGroup(msg.ChatID, groupAlias, users)
	if err != nil {
		log.Printf("[WARN] error while adding group alias %s:%s: %+v", msg.ChatID, groupAlias, err)
		if g.RespondAllCommands {
			return &Response{
				Reply: true,
				Text:  "Internal error",
			}
		}
		return nil
	}

	return &Response{Reply: true, Text: fmt.Sprintf("Group %s has been successfully added", groupAlias)}
}

// Help returns the usage of this bot
func (g *GroupBot) Help() string {
	return `/add_group @<group alias> @<user1>, @<user2>, ... - add user
/delete_user_from_group @<group alias> @<user> - removes user from the group
/detete_group @<group alias> - removes group
/list_groups - shows the list of existing groups
/add_user_to_group @<group alias> @<user> - adds user to the specified group
@<group alias> - triggers bot to send message with all participants of the group`
}

// prepareIllegalAccessMessage creates a response to the illegal
// command execution - if in bot parameters defined to respond all
// commands - it will return a message, otherwise - nothing
func (g *GroupBot) prepareIllegalAccessMessage() *Response {
	if g.RespondAllCommands {
		return &Response{Reply: true, Text: "You don't have admin rights to execute this command"}
	}
	return nil
}

// unique returns slice of unique string occurrences from the source one
func unique(sl []string) []string {
	m := make(map[string]struct{})
	for _, s := range sl {
		m[s] = struct{}{}
	}
	var res []string
	for s := range m {
		res = append(res, s)
	}
	return res
}
