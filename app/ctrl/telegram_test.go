package ctrl

import (
	"context"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Semior001/multibot-utility/app/bot"
)

func TestTelegramBotCtrl_RunWithNoBots(t *testing.T) {
	defer checkPanics(t)
	bots := bot.MockBot{}
	api := mockTbAPI{}
	ctrl := TelegramBotCtrl{
		Token:    "",
		Bots:     &bots,
		API:      &api,
		UserName: "",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	updMsg := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{},
			Text: "12345",
		},
	}

	api.On("GetChat")
	updChan := make(chan tgbotapi.Update, 1)
	updChan <- updMsg
	close(updChan)

	api.On("GetUpdatesChan", mock.Anything).Return(tgbotapi.UpdatesChannel(updChan), nil)
	bots.On("OnMessage", mock.Anything).Return(nil)
	err := ctrl.Run(ctx)
	assert.EqualError(t, err, "telegram updates chan closed")
}

func TestTelegramBotCtrl_convertMessage(t *testing.T) {
	defer checkPanics(t)
	api := mockTbAPI{}
	api.On("GetChatAdministrators", mock.MatchedBy(func(config tgbotapi.ChatConfig) bool {
		return config.ChatID == 555
	})).Return([]tgbotapi.ChatMember{
		{
			User: &tgbotapi.User{
				ID:           999,
				FirstName:    "Yelshat",
				LastName:     "Duskaliyev",
				UserName:     "semior001",
				LanguageCode: "ru",
			},
		},
	}, nil)
	api.On("GetChatAdministrators", mock.Anything).Return(nil, nil)
	ctrl := TelegramBotCtrl{
		API:      &api,
		UserName: "testbot",
	}

	msg := &tgbotapi.Message{
		MessageID: 123,
		Chat:      &tgbotapi.Chat{ID: 321, Type: "group"},
		From: &tgbotapi.User{
			ID:        456,
			FirstName: "Yelshat",
			LastName:  "Duskaliyev",
			UserName:  "semior001",
			IsBot:     false,
		},
		Date: 1587732716,
		Text: "abc",
	}
	expected := bot.Message{
		ID:       "123",
		ChatID:   "321",
		ChatType: bot.ChatTypeGroup,
		From: &bot.User{
			ID:          "456",
			Username:    "semior001",
			DisplayName: "Yelshat Duskaliyev",
		},
		Sent:           time.Unix(1587732716, 0),
		Text:           "abc",
		AddedBotToChat: false,
	}
	assert.Equal(t, expected, ctrl.convertMessage(msg), "group chat")

	msg.Chat.Type = "private"
	expected.ChatType = bot.ChatTypePrivate
	assert.Equal(t, expected, ctrl.convertMessage(msg), "private chat")

	msg.Chat.Type = "channel"
	expected.ChatType = bot.ChatTypeChannel
	assert.Equal(t, expected, ctrl.convertMessage(msg), "channel chat")

	msg = &tgbotapi.Message{
		MessageID: 125,
		Chat:      &tgbotapi.Chat{ID: 322, Type: "group"},
		NewChatMembers: &[]tgbotapi.User{
			{
				ID:       222,
				UserName: "testbot",
				IsBot:    true,
			},
		},
		Date: 1587732716,
	}
	expected = bot.Message{
		ID:             "125",
		ChatID:         "322",
		ChatType:       bot.ChatTypeGroup,
		Sent:           time.Unix(1587732716, 0),
		Text:           "",
		AddedBotToChat: true,
	}
	assert.Equal(t, expected, ctrl.convertMessage(msg), "added bot to chat")

	expected = bot.Message{
		ID:       "555",
		ChatID:   "666",
		ChatType: bot.ChatTypeGroup,
		Sent:     time.Unix(1587732716, 0),
		Text:     "/start",
		From: &bot.User{
			ID:          "456",
			Username:    "semior001",
			DisplayName: "Yelshat Duskaliyev",
		},
		AddedBotToChat: true,
	}
	msg = &tgbotapi.Message{
		MessageID: 555,
		Chat:      &tgbotapi.Chat{ID: 666, Type: "group"},
		From: &tgbotapi.User{
			ID:        456,
			FirstName: "Yelshat",
			LastName:  "Duskaliyev",
			UserName:  "semior001",
			IsBot:     false,
		},
		Date: 1587732716,
		Text: "/start",
	}
	assert.Equal(t, expected, ctrl.convertMessage(msg), "added bot to chat with /start")

	msg = &tgbotapi.Message{
		MessageID: 123,
		Chat: &tgbotapi.Chat{
			ID:                  321,
			Type:                "group",
			AllMembersAreAdmins: true,
		},
		From: &tgbotapi.User{
			ID:        456,
			FirstName: "Yelshat",
			LastName:  "Duskaliyev",
			UserName:  "semior001",
			IsBot:     false,
		},
		Date: 1587732716,
		Text: "abc",
	}
	expected = bot.Message{
		ID:       "123",
		ChatID:   "321",
		ChatType: bot.ChatTypeGroup,
		From: &bot.User{
			ID:          "456",
			Username:    "semior001",
			DisplayName: "Yelshat Duskaliyev",
			IsAdmin:     true,
		},
		Sent:           time.Unix(1587732716, 0),
		Text:           "abc",
		AddedBotToChat: false,
	}
	assert.Equal(t, expected, ctrl.convertMessage(msg), "all members are admins")

	msg = &tgbotapi.Message{
		MessageID: 123,
		Chat: &tgbotapi.Chat{
			ID:                  555,
			Type:                "group",
			AllMembersAreAdmins: false,
		},
		From: &tgbotapi.User{
			ID:        999,
			FirstName: "Yelshat",
			LastName:  "Duskaliyev",
			UserName:  "semior001",
			IsBot:     false,
		},
		Date: 1587732716,
		Text: "abc",
	}
	expected = bot.Message{
		ID:       "123",
		ChatID:   "555",
		ChatType: bot.ChatTypeGroup,
		From: &bot.User{
			ID:          "999",
			Username:    "semior001",
			DisplayName: "Yelshat Duskaliyev",
			IsAdmin:     true,
		},
		Sent:           time.Unix(1587732716, 0),
		Text:           "abc",
		AddedBotToChat: false,
	}

	transform := ctrl.convertMessage(msg)
	if !reflect.DeepEqual(expected, transform) {
		t.Errorf("api request to get admins \n expected: \n %+v \n got: \n %+v", expected, transform)
	}
}

func TestTelegramBotCtrl_sendBotResponse(t *testing.T) {
	api := mockTbAPI{}
	ctrl := TelegramBotCtrl{
		API: &api,
	}

	api.On("Send", mock.Anything).Return(tgbotapi.Message{
		MessageID: 5555,
	}, nil)
	api.On("PinChatMessage", mock.Anything).Return(tgbotapi.APIResponse{
		Ok:          true,
		Result:      nil,
		ErrorCode:   0,
		Description: "",
		Parameters:  nil,
	}, nil)
	api.On("UnpinChatMessage", mock.Anything).Return(tgbotapi.APIResponse{
		Ok:          true,
		Result:      nil,
		ErrorCode:   0,
		Description: "",
		Parameters:  nil,
	}, nil)

	// if we want to pin and unping message consequently - we just want to ping all
	// users in the chat
	err := ctrl.SendBotResponse(&bot.Response{
		Pin:   true,
		Unpin: true,
	}, "1234")
	require.NoError(t, err)
}

func checkPanics(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("Caught panic: \n %+v \n stacktrace: \n %+v", r, string(debug.Stack()))
	}
}
