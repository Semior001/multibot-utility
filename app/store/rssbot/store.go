package rssbot

//go:generate mockery -inpkg -name Store -case snake

// Store defines methods to store and fetch
// RSS feeds and corresponding chats
type Store interface {
	AddRSSFeed(chatID string, URL string) (err error)
	AddChat(id string) (err error)
}
