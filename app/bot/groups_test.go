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
	require.Equal(t, `/add_group @<group alias> @<user1>, @<user2>, ... - add user
/delete_user_from_group @<group alias> @<user> - removes user from the group
/detete_group @<group alias> - removes group
/list_groups - shows the list of existing groups
/add_user_to_group @<group alias> @<user> - adds user to the specified group
@<group alias> - triggers bot to send message with all participants of the group`, (&GroupBot{}).Help())
}

func TestGroupBot_Add(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	var storeInterface groups.Store = &mockGroupStore

	b := NewGroupBot(&storeInterface)

	resp := b.OnMessage(Message{
		From: User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @admins @test @test1 @test2 @test3",
	})

	assert.Equal(t, "Group @admins has been successfully added", resp.Text)
}

func TestGroupBot_List(t *testing.T) {
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

	var storeInterface groups.Store = &mockGroupStore

	b := NewGroupBot(&storeInterface)

	for i := 0; i < 10; i++ {
		b.OnMessage(Message{
			From: User{
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

	assert.Equal(t, `@admins_0 : @test, @test1, @test2, @test3
@admins_1 : @test, @test1, @test2, @test3
@admins_2 : @test, @test1, @test2, @test3
@admins_3 : @test, @test1, @test2, @test3
@admins_4 : @test, @test1, @test2, @test3
@admins_5 : @test, @test1, @test2, @test3
@admins_6 : @test, @test1, @test2, @test3
@admins_7 : @test, @test1, @test2, @test3
@admins_8 : @test, @test1, @test2, @test3
@admins_9 : @test, @test1, @test2, @test3`, resp.Text)
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

	var storeInterface groups.Store = &mockGroupStore

	b := NewGroupBot(&storeInterface)

	b.OnMessage(Message{
		From: User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @admins @test @test1 @test2 @test3",
	})

	resp := b.OnMessage(Message{
		From: User{
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

	var storeInterface groups.Store = &mockGroupStore

	b := NewGroupBot(&storeInterface)

	b.OnMessage(Message{
		From: User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @admins @test @test1 @test2 @test3",
	})

	resp := b.OnMessage(Message{
		From: User{
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

func TestGroupBot_TestTrigger(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	mockGroupStore.On(
		"PutGroup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)
	mockGroupStore.On(
		"GetGroup",
		"@some_students",
	).Return([]string{
		"@blah",
		"@blah1",
		"@blah2",
	}, nil)

	var storeInterface groups.Store = &mockGroupStore

	b := NewGroupBot(&storeInterface)

	b.OnMessage(Message{
		From: User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     true,
		},
		Text: "/add_group @some_students @blah @blah1 @blah2",
	})

	resp := b.OnMessage(Message{
		Text: "There is a reference to @some_students",
	})
	assert.Equal(t, "@blah @blah1 @blah2", resp.Text)
}

func TestGroupBot_FailedAuth(t *testing.T) {
	mockGroupStore := groups.MockStore{}
	var storeInterface groups.Store = &mockGroupStore
	b := NewGroupBot(&storeInterface)

	resp := b.OnMessage(Message{
		From: User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/add_group @some_students @blah @blah1 @blah2",
	})
	assert.Equal(t, nil, resp)

	resp = b.OnMessage(Message{
		From: User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/delete_group @some_students",
	})
	assert.Equal(t, nil, resp)

	resp = b.OnMessage(Message{
		From: User{
			Username:    "blah",
			DisplayName: "blahblah",
			IsAdmin:     false,
		},
		Text: "/delete_user_from_group @some_students @blah",
	})
	assert.Equal(t, nil, resp)
}

func TestGroupBot_AddUser(t *testing.T) {
	panic("todo")
}
