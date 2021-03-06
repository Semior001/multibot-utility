package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Semior001/multibot-utility/app/store/groups"
)

const regexpAlias = "@[a-zA-Z0-9_]+"
const aliasPrefix = "@"

// todo limit on message length

// GroupBotParams describes all necessary parameters for correct working of GroupBot
type GroupBotParams struct {
	Store              groups.Store
	RespondAllCommands bool
	GetGroupMembers    func(chatID string) ([]User, error)
}

// GroupBot gathers usernames into one mention, like @admins
type GroupBot struct {
	GroupBotParams
}

// NewGroupBot initializes an instance of GroupBot
func NewGroupBot(params GroupBotParams) *GroupBot {
	log.Print("[INFO] GroupBot instantiated")
	return &GroupBot{
		GroupBotParams: params,
	}
}

// OnMessage receives any commands, that are listed in help and group aliases
func (g *GroupBot) OnMessage(msg Message) *Response {
	// ignore all non-group messages
	if msg.ChatType != ChatTypeGroup {
		return nil
	}

	// if bot has been added to chat, we have to save this chat in the storage
	if msg.AddedBotToChat {
		if err := g.Store.AddChat(msg.ChatID); err != nil {
			log.Printf("[WARN] error while adding chat to store: %+v", err)
		}
		return nil
	}

	trimmed := removeRedundantWhitespaces(msg.Text)

	tokens := strings.Split(trimmed, " ")

	// command may be in format /cmd@bot
	// todo if user wanted to ping other bot - we should not react
	cmd := strings.Split(tokens[0], aliasPrefix)[0]
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

	// checking whether the alias is @all and we have a callback
	// to get group members
	if g.GetGroupMembers != nil && contains(aliases, "@all") {
		// adding everyone into message
		users, err := g.GetGroupMembers(msg.ChatID)
		if err != nil {
			log.Printf("[WARN] failed to get group members after trigger @all %+v", err)
			return nil
		}
		resp := strings.Builder{}

		for _, u := range users {
			if !u.IsBot {
				_, _ = resp.WriteString(strings.ReplaceAll(u.Username, "_", "\\_") + " ")
			}
		}

		// composing users into one ping message
		return &Response{Text: resp.String()}
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

	resp := strings.Builder{}

	for _, u := range users {
		_, _ = resp.WriteString(escapeUnderscores(u) + " ")
	}

	// composing users into one ping message
	return &Response{Text: resp.String()}
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

	if !strings.HasPrefix(user, aliasPrefix) {
		user = aliasPrefix + user
	}

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
		Text:  fmt.Sprintf("User %s has been successfully added to the group %s", removeUsersPings(user), groupAlias),
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
		groupStrings = append(groupStrings, fmt.Sprintf("%s : %s", alias,
			removeUsersPings(escapeUnderscores(
				strings.Join(users, ", "),
			)),
		))
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
	return &Response{Reply: true, Text: fmt.Sprintf("Group %s has been successfully deleted", groupAlias)}
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

	return &Response{
		Reply: true,
		Text:  fmt.Sprintf("User %s has been successfully deleted from group %s", removeUsersPings(user), groupAlias),
	}
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

	for i := range users {
		if !strings.HasPrefix(users[i], aliasPrefix) {
			users[i] = aliasPrefix + users[i]
		}
	}

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
	return `Groups bot - gathers usernames into one mention, like @admins
/add\_group @group\_alias @user1, @user2, ... - add user
/delete\_user\_from\_group @group\_alias @user - removes user from the group
/delete\_group @group\_alias - removes group
/list\_groups - shows the list of existing groups
/add\_user\_to\_group @group\_alias @user - adds user to the specified group
@group\_alias - triggers bot to send message with all participants of the group`
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

// removeUsersPings removes all aliasPrefix occurrences from string to not ping user in chat
func removeUsersPings(s string) string {
	return strings.ReplaceAll(s, aliasPrefix, "")
}

// escapeUnderscores escapes all underscores in the given string
func escapeUnderscores(s string) string {
	return strings.ReplaceAll(s, "_", "\\_")
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

// remove all excess whitespaces from string
func removeRedundantWhitespaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
