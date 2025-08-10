package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	models "tele-goat-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var randomWordArray []string
var chatStore = models.ChatStore{Chats: make(map[int64]models.Chat)}

// Filters out service messages (join, leave, pin, etc.)
func isTextMessage(m *tgbotapi.Message) bool {
	if m == nil {
		return false
	}
	if m.IsCommand() {
		return true
	}
	if m.Text == "" {
		return false
	}
	if m.From == nil {
		return false
	}
	return true
}

// Checks if user guessed the current game word
func isRight(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(strings.ToLower(update.Message.Text))

	if chat, ok := chatStore.Chats[chatID]; ok && text == strings.ToLower(strings.TrimSpace(chat.SpeedWord)) {
		announce(bot, chatID, update.Message, "%s got it first!")
		chat.SpeedWord = ""
		chatStore.UpdateChat(chatID, chat)
		return
	}
	if chat, ok := chatStore.Chats[chatID]; ok && text == strings.ToLower(strings.TrimSpace(chat.ScrambledWord)) {
		announce(bot, chatID, update.Message, "%s unscrambled the word")
		chat.ScrambledWord = ""
		chatStore.UpdateChat(chatID, chat)
		return
	}
	if chat, ok := chatStore.Chats[chatID]; ok && text == strings.ToLower(strings.TrimSpace(chat.TabooWord)) {
		announce(bot, chatID, update.Message, "%s guessed the word correctly")
		chat.TabooWord = ""
		chatStore.UpdateChat(chatID, chat)
		return
	}
}
func main() {
	rand.Seed(time.Now().UnixNano())

	// Load words from file, trim, filter empties
	f, err := os.Open("words.txt")
	if err != nil {
		log.Fatal(err)
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w != "" {
			randomWordArray = append(randomWordArray, w)
		}
	}
	_ = f.Close()
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
	if len(randomWordArray) == 0 {
		log.Fatal("words.txt has no usable words")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELE_GOAT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = os.Getenv("BOT_DEBUG") == "1"

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Drain and advance offset so we don't replay old updates
	if ups, err := bot.GetUpdates(u); err == nil && len(ups) > 0 {
		u.Offset = ups[len(ups)-1].UpdateID + 1
	}

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		m := update.Message
		if m == nil {
			continue
		}

		// Commands
		if m.IsCommand() {
			handleCommand(bot, update)
			continue
		}

		// Regular messages
		if !isTextMessage(m) {
			continue
		}

		isRight(bot, update)
	}
}
