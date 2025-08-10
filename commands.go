package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	models "tele-goat-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Command router
func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	switch strings.ToLower(update.Message.Command()) {
	case "help":
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "No help"))
	case "word":
		wordCommand(bot, update)
	case "scramble":
		scrambleCommand(bot, update)
	case "taboo":
		tabooCommand(bot, update)
	case "fscramble":
		forfeitScrambleCommand(bot, update)
	case "ftaboo":
		forfeitTabooCommand(bot, update)
	default:
		_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Unknown command"))
	}
}

// Helper to send messages about who guessed first
func announce(bot *tgbotapi.BotAPI, chatID int64, m *tgbotapi.Message, tmpl string) {
	if m.From == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(tmpl, m.From.FirstName))
	msg.ReplyToMessageID = m.MessageID
	if _, err := bot.Send(msg); err != nil {
		log.Println("send:", err)
	}
}

// Picks a random word from the loaded list
func randomWord() string {
	return randomWordArray[rand.Intn(len(randomWordArray))]
}

// === Command handlers ===

func wordCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	w := randomWord()
	chatID := update.Message.Chat.ID

	if !chatStore.ChatFound(chatID) {
		chat := models.Chat{ID: chatID, SpeedWord: w}
		chatStore.AddChat(chat)
	} else {
		chat := chatStore.Chats[chatID]
		chat.SpeedWord = w
		chatStore.UpdateChat(chatID, chat)
	}

	_, _ = bot.Send(tgbotapi.NewMessage(chatID, "Type: "+w))
}

func scrambleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	w := randomWord()
	chatID := update.Message.Chat.ID

	if chatStore.ChatFound(chatID) && chatStore.Chats[chatID].ScrambledWord != "" {
		msg := tgbotapi.NewMessage(chatID, "Word already in play. Please use the /fscramble command")
		msg.ReplyToMessageID = update.Message.MessageID
		_, _ = bot.Send(msg)
		return
	}

	// Rune-safe scramble
	r := []rune(w)
	rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
	scrambled := string(r)

	if !chatStore.ChatFound(chatID) {
		chat := models.Chat{ID: chatID, ScrambledWord: w}
		chatStore.AddChat(chat)
	} else {
		chat := chatStore.Chats[chatID]
		chat.ScrambledWord = w
		chatStore.UpdateChat(chatID, chat)
	}

	_, _ = bot.Send(tgbotapi.NewMessage(chatID, "unscramble: "+scrambled))
}

func forfeitScrambleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	if !chatStore.ChatFound(chatID) {
		chatStore.AddChat(models.Chat{ID: chatID})
		return
	}

	chat := chatStore.Chats[chatID]
	w := chat.ScrambledWord
	chat.ScrambledWord = ""
	chatStore.UpdateChat(chatID, chat)

	_, _ = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("The scrambled word was %s", w)))
}

func tabooCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message.From == nil {
		return
	}

	w := randomWord()
	chatID := update.Message.Chat.ID
	user := update.Message.From

	if chatStore.ChatFound(chatID) && chatStore.Chats[chatID].TabooWord != "" {
		msg := tgbotapi.NewMessage(chatID, "Word already in play. Please use the /ftaboo command")
		msg.ReplyToMessageID = update.Message.MessageID
		_, _ = bot.Send(msg)
		return
	}

	// PM the user
	tabooWordMsg := tgbotapi.NewMessage(int64(user.ID), fmt.Sprintf("Your word is \"%s\" for chat: %s", w, update.Message.Chat.Title))
	_, err := bot.Send(tabooWordMsg)

	if err != nil {
		errMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Hey %s, I can't privately message you unless you talk to me first", user.FirstName))
		_, _ = bot.Send(errMsg)
		return
	}

	if !chatStore.ChatFound(chatID) {
		chat := models.Chat{ID: chatID, TabooWord: w, TabooUserID: user.ID}
		chatStore.AddChat(chat)
	} else {
		chat := chatStore.Chats[chatID]
		chat.TabooWord = w
		chat.TabooUserID = user.ID
		chatStore.UpdateChat(chatID, chat)
	}

	confirmationMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Word has been sent.\n%s will now describe the word without saying it.", user.FirstName))
	_, _ = bot.Send(confirmationMsg)
}

func forfeitTabooCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	if !chatStore.ChatFound(chatID) {
		chatStore.AddChat(models.Chat{ID: chatID})
		return
	}

	chat := chatStore.Chats[chatID]
	w := chat.TabooWord
	chat.TabooWord, chat.TabooUserID = "", 0
	chatStore.UpdateChat(chatID, chat)

	_, _ = bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("The word that was sent out was %s", w)))
}
