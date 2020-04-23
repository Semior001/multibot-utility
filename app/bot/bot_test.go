package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMultiBot_Help(t *testing.T) {
	mockBot := MockBot{}
	mockBot.On("Help", mock.Anything).Return("blahblahblah")
	bot := MultiBot{
		&mockBot,
	}
	assert.Equal(t, "blahblahblah", bot.Help())
	assert.Equal(t, &Response{
		Text: "blahblahblah",
	}, bot.OnMessage(Message{
		Text: "help",
	}))
}

func TestMultiBot_OnMessage(t *testing.T) {
	mockBot := MockBot{}
	mockBot.On("OnMessage", mock.Anything).Return(&Response{
		Text:        "foo",
		Pin:         true,
		Unpin:       true,
		Preview:     true,
		Reply:       true,
		BanInterval: 999,
	})
	bot := MultiBot{
		&mockBot,
	}
	assert.Equal(t, &Response{
		Text:        "foo",
		Pin:         true,
		Unpin:       true,
		Preview:     true,
		Reply:       true,
		BanInterval: 999,
	}, bot.OnMessage(Message{
		Text: "blahblah",
	}))
}
