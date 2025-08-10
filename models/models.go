package models

type Chat struct {
	ID            int64
	TabooWord     string
	TabooUserID   int // was int
	ScrambledWord string
	SpeedWord     string
}

type ChatStore struct {
	Chats map[int64]Chat
}

func (s *ChatStore) ChatFound(chatID int64) bool {
	_, ok := s.Chats[chatID]
	return ok
}

func (s *ChatStore) Get(chatID int64) (Chat, bool) {
	c, ok := s.Chats[chatID]
	return c, ok
}

func (s *ChatStore) AddChat(chat Chat) bool {
	if _, exists := s.Chats[chat.ID]; exists {
		return false
	}
	s.Chats[chat.ID] = chat
	return true
}

func (s *ChatStore) UpdateChat(chatID int64, updated Chat) bool {
	if _, ok := s.Chats[chatID]; !ok {
		return false
	}
	s.Chats[chatID] = updated
	return true
}
