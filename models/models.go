package models

type ChatStore struct {
	Chats map[int64]Chat
}

func (store ChatStore) ChatFound(chatID int64) bool {

	if _, ok := store.Chats[chatID]; ok {
		return true
	}
	return false
}

func (store ChatStore) AddChat(chat Chat) {
	if !store.ChatFound(chat.ID) {
		store.Chats[chat.ID] = chat
	}
}

func (store ChatStore) UpdateChat(chatID int64, updatedChat Chat) {
	if _, ok := store.Chats[chatID]; ok {
		store.Chats[chatID] = updatedChat
	}
}

type Chat struct {
	ID            int64
	TabooWord     string
	TabooUserID   int
	ScrambledWord string
	SpeedWord     string
}
