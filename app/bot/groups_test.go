package bot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/Semior001/multibotUtility/app/store/groups"

	"github.com/stretchr/testify/require"
)

func TestGroupBot_Help(t *testing.T) {
	require.Equal(t, `/add\_group @group\_alias @user1, @user2, ... - add user
/delete\_user\_from\_group @group\_alias @user - removes user from the group
/detete\_group @group\_alias - removes group
/list\_groups - shows the list of existing groups
/add\_user\_to\_group @group\_alias @user - adds user to the specified group
@group\_alias - triggers bot to send message with all participants of the group`, (&GroupBot{}).Help())
}

func TestGroupBot_AddGroup(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	resp := b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @admins @test @test1 @test2 @test3",
	})

	assert.Equal(t, "Group @admins has been successfully added", resp.Text)
}

func TestGroupBot_ListGroups(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	mockGroupStore.On(
		"GetGroups",
		mock.Anything,
	).Return(map[string][]string{
		"@admins_0": {"@test", "@test1", "@test2", "@test3"},
		"@admins_1": {"@test", "@test1", "@test2", "@test3"},
		"@admins_2": {"@test", "@test1", "@test2", "@test3"},
		"@admins_3": {"@test", "@test1", "@test2", "@test3"},
		"@admins_4": {"@test", "@test1", "@test2", "@test3"},
		"@admins_5": {"@test", "@test1", "@test2", "@test3"},
		"@admins_6": {"@test", "@test1", "@test2", "@test3"},
		"@admins_7": {"@test", "@test1", "@test2", "@test3"},
		"@admins_8": {"@test", "@test1", "@test2", "@test3"},
		"@admins_9": {"@test", "@test1", "@test2", "@test3"},
	}, nil)

	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	for i := 0; i < 10; i++ {
		b.OnMessage(Message{
			From: &User{
				Username:    "blah",
				DisplayName: "blahblah",
				IsAdmin:     true,
			},
			Text: fmt.Sprintf("/add_group @admins_%d @test @test1 @test2 @test3", i),
		})
	}

	resp := b.OnMessage(Message{
		Text: "/list_groups",
	})

	for i := 0; i < 10; i++ {
		assert.Contains(t, resp.Text, fmt.Sprintf("@admins_%d : @test, @test1, @test2, @test3", i))
	}
}

func TestGroupBot_DeleteUserFromGroup(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	mockGroupStore.On(
		"DeleteUserFromGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @admins @test @test1 @test2 @test3",
	})

	resp := b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/delete_user_from_group @admins @test1",
	})

	assert.Equal(t, "User @test1 has been successfully deleted from group @admins", resp.Text)
}

func TestGroupBot_DeleteGroup(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	mockGroupStore.On(
		"DeleteGroup",
		mock.Anything,
		mock.Anything,
	).Return(nil)
	mockGroupStore.On(
		"GetGroups",
		mock.Anything,
	).Return(map[string][]string{}, nil)

	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @admins @test @test1 @test2 @test3",
	})

	resp := b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/delete_group @admins",
	})
	assert.Equal(t, "Group @admins has been successfully deleted", resp.Text)

	resp = b.OnMessage(Message{
		Text: "/list_groups",
	})
	assert.Equal(t, "There's no groups in this chat yet", resp.Text)
}

func TestGroupBot_AddUser(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"AddUser",
		"",
		"@some_students",
		"@blah",
	).Return(nil)
	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	resp := b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_user_to_group @some_students @blah",
	})

	assert.Equal(t, "User @blah has been successfully added to the group @some_students", resp.Text)
}

func TestGroupBot_Trigger(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	mockGroupStore.On(
		"FindAliases",
		mock.Anything, []string{"@some_students", "@kek"},
	).Return([]string{"@blah", "@blah1", "@blah2", "@blah3", "@blah4"}, nil)
	mockGroupStore.On(
		"FindAliases",
		mock.Anything, []string{"@kek", "@some_students"},
	).Return([]string{"@blah", "@blah1", "@blah2", "@blah3", "@blah4"}, nil)

	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @some_students @blah @blah1 @blah2",
	})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @kek @blah @blah3 @blah4",
	})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @lol @blah5 @blah6 @blah7",
	})

	resp := b.OnMessage(Message{
		Text: "There is a reference to @some_students and @kek",
	})

	assert.Contains(t, resp.Text, "@blah")
	assert.Contains(t, resp.Text, "@blah1")
	assert.Contains(t, resp.Text, "@blah2")
	assert.Contains(t, resp.Text, "@blah3")
	assert.Contains(t, resp.Text, "@blah4")
	assert.NotContains(t, resp.Text, "@blah5")
	assert.NotContains(t, resp.Text, "@blah6")
	assert.NotContains(t, resp.Text, "@blah7")
}

func TestGroupBot_Unique(t *testing.T) {
	queried := unique([]string{"@blah", "@blah1", "@blah", "@blah1", "@blah3"})
	m := make(map[string]int)
	for _, s := range queried {
		if _, ok := m[s]; !ok {
			m[s] = 0
		}
		m[s]++
	}
	for _, cnt := range m {
		assert.Equal(t, cnt, 1)
	}
}

func TestGroupBot_TriggerNoAliases(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	mockGroupStore.On(
		"FindAliases",
		mock.Anything, mock.Anything,
	).Return([]string{}, nil)

	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @some_students @blah @blah1 @blah2",
	})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @kek @blah @blah3 @blah4",
	})

	b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @lol @blah5 @blah6 @blah7",
	})

	resp := b.OnMessage(Message{
		Text: "There is a reference to nobody",
	})
	assert.Nil(t, resp)
}

func TestGroupBot_IllegalArgumentsNumber(t *testing.T) {
	// add user
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"AddUser",
		"",
		"@some_students",
		"@blah",
	).Return(nil)
	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: true})

	resp := b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_user_to_group @some_students",
	})

	assert.Equal(t, "Command requires exactly two arguments - group alias and username", resp.Text)

	// delete user from group
	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/delete_user_from_group @some_students",
	})

	assert.Equal(t, "Command requires exactly two arguments - group alias and username", resp.Text)

	// delete group
	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/delete_group",
	})

	assert.Equal(t, "Command requires exactly one argument - group alias", resp.Text)

	// add group
	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @blah",
	})

	assert.Equal(t, "Not enough parameters to add group", resp.Text)

	// without responding
	b = NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_user_to_group @some_students",
	})

	assert.Nil(t, resp)

	// delete user from group
	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/delete_user_from_group @some_students",
	})

	assert.Nil(t, resp)

	// delete group
	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/delete_group",
	})

	assert.Nil(t, resp)

	// add group
	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @blah",
	})

	assert.Nil(t, resp)

}

func TestGroupBot_IllegalAccess(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})

	resp := b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/add_group @some_students @blah @blah1 @blah2",
	})
	assert.Equal(t, (*Response)(nil), resp)

	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/delete_group @some_students",
	})
	assert.Equal(t, (*Response)(nil), resp)

	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/delete_user_from_group @some_students @blah",
	})
	assert.Equal(t, (*Response)(nil), resp)

	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/add_user_to_group @some_students @blah",
	})
	assert.Equal(t, (*Response)(nil), resp)

	// with responding
	b = NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: true})
	resp = b.OnMessage(Message{
		From: &User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/add_user_to_group @some_students @blah",
	})
	assert.Equal(t, "You don't have admin rights to execute this command", resp.Text)

}

func TestGroupBot_AddChat(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On("AddChat", mock.Anything).Return(nil)

	b := NewGroupBot(GroupBotParams{Store: &mockGroupStore, RespondAllCommands: false})
	resp := b.OnMessage(Message{
		ChatID:         "qwerty",
		AddedBotToChat: true,
	})
	assert.Nil(t, resp)
}

func TestGroupBot_TriggerAll(t *testing.T) {
	// add user
	mockGroupStore := groups.MockStore{}
	b := NewGroupBot(GroupBotParams{
		Store:              &mockGroupStore,
		RespondAllCommands: false,
		GetGroupMembers: func(chatID string) (users []User, err error) {
			return []User{
				{
					Username: "@semior001",
					IsBot:    false,
				},
				{
					Username: "@blah",
					IsBot:    false,
				},
				{
					Username: "@sMultibot",
					IsBot:    true,
				},
				{
					Username: "@blah1",
					IsBot:    false,
				},
			}, nil
		},
	})
	resp := b.OnMessage(Message{
		Text: "There is a reference to @all",
	})

	assert.Contains(t, resp.Text, "@semior001")
	assert.Contains(t, resp.Text, "@blah")
	assert.Contains(t, resp.Text, "@blah1")
	assert.NotContains(t, resp.Text, "@sMultibot")
}
