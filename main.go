package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	s "strings"
	"time"

	models "tele-goat-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var wordDict = make(map[int64]string)
var scrambleDict = make(map[int64]string)
var tabooDict = make(map[int64]string)
var randomWordArray []string

var chatStore = models.ChatStore{Chats: make(map[int64]models.Chat)}

func main() {

	content, err := ioutil.ReadFile("words.txt")

	if err != nil {
		log.Panic(err)
	}

	randomWordArray = s.Split(string(content), "\n")

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELE-GOAT-SECRET"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch s.ToLower(update.Message.Command()) {
			case "help":
				msg.Text = "No help"

			case "word":
				wordCommand(bot, update)

			case "scramble":
				scrambleCommand(bot, update)

			case "taboo":
				tabooCommand(bot, update)

			case "ftaboo":
				forfeitTabooCommand(bot, update)

			case "fscramble":
				forfeitScrambleCommand(bot, update)
			}

			bot.Send(msg)
		}

		isRight(bot, update)

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

	}
}

func isRight(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if chat, ok := chatStore.Chats[update.Message.Chat.ID]; ok {
		if s.TrimSpace(s.ToLower(update.Message.Text)) == s.TrimSpace(s.ToLower(chat.SpeedWord)) {
			chatID := update.Message.Chat.ID

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("%s got it first!", update.Message.From.FirstName))
			msg.ReplyToMessageID = update.Message.MessageID
			chat.SpeedWord = ""
			chatStore.UpdateChat(chat.ID, chat)
			bot.Send(msg)
		}
	}

	if chat, ok := chatStore.Chats[update.Message.Chat.ID]; ok {
		if s.TrimSpace(s.ToLower(update.Message.Text)) == s.TrimSpace(s.ToLower(chat.ScrambledWord)) {
			chatID := update.Message.Chat.ID

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("%s unscrambled the word", update.Message.From.FirstName))
			msg.ReplyToMessageID = update.Message.MessageID
			chat.ScrambledWord = ""
			chatStore.UpdateChat(chatID, chat)
			bot.Send(msg)
		}
	}

	if chat, ok := chatStore.Chats[update.Message.Chat.ID]; ok {
		if s.TrimSpace(s.ToLower(update.Message.Text)) == s.TrimSpace(s.ToLower(chat.TabooWord)) {
			chatID := update.Message.Chat.ID

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("%s guessed the word correctly", update.Message.From.FirstName))
			msg.ReplyToMessageID = update.Message.MessageID
			chat.TabooWord = ""
			chatStore.UpdateChat(chatID, chat)
			bot.Send(msg)
		}
	}
}

func tabooCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	tabooRandomWord := randomWord()
	chatID := update.Message.Chat.ID
	user := update.Message.From

	if chatStore.ChatFound(chatID) && chatStore.Chats[chatID].TabooWord != "" {

		msg := tgbotapi.NewMessage(chatID, "Word already in play. Please use the /ftaboo command")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	tabooWordMsg := tgbotapi.NewMessage((int64(user.ID)), fmt.Sprintf("Your word is \"%s\" for chat: %s", tabooRandomWord, update.Message.Chat.Title))
	_, err := bot.Send(tabooWordMsg)

	if err != nil {
		errMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Hey %s, I can't privately message you unless you talk to me first", user.FirstName))
		bot.Send(errMsg)

	} else {

		if !chatStore.ChatFound(update.Message.Chat.ID) {
			chat := models.Chat{
				ID:          update.Message.Chat.ID,
				TabooWord:   tabooRandomWord,
				TabooUserID: update.Message.From.ID,
			}
			chatStore.AddChat(chat)

		} else {
			chat := chatStore.Chats[chatID]
			chat.TabooWord = tabooRandomWord
			chat.TabooUserID = update.Message.From.ID
			chatStore.UpdateChat(chatID, chat)
		}

		confirmationMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Word have been sent.\n%s will now describe the word without saying it.", user.FirstName))
		bot.Send(confirmationMsg)
	}

}

func wordCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	randomWord := randomWord()
	chatID := update.Message.Chat.ID

	if !chatStore.ChatFound(update.Message.Chat.ID) {
		chat := models.Chat{
			ID:        update.Message.Chat.ID,
			SpeedWord: randomWord,
		}
		chatStore.AddChat(chat)
	} else {
		chat := chatStore.Chats[chatID]
		chat.SpeedWord = randomWord
		chatStore.UpdateChat(chatID, chat)
	}

	msg := tgbotapi.NewMessage(chatID, "Type: "+chatStore.Chats[chatID].SpeedWord)
	bot.Send(msg)
}

func scrambleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	rand.Seed(time.Now().UnixNano())
	scrambleRandomWord := randomWord()
	chatID := update.Message.Chat.ID

	if chatStore.ChatFound(chatID) && chatStore.Chats[chatID].ScrambledWord != "" {

		msg := tgbotapi.NewMessage(chatID, "Word already in play. Please use the /fscramble command")
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		return
	}

	scrambledCharList := s.Split(scrambleRandomWord, "")
	rand.Shuffle(len(scrambledCharList), func(i, j int) {
		scrambledCharList[i], scrambledCharList[j] = scrambledCharList[j], scrambledCharList[i]
	})

	if !chatStore.ChatFound(update.Message.Chat.ID) {
		chat := models.Chat{
			ID:            update.Message.Chat.ID,
			ScrambledWord: scrambleRandomWord,
		}
		chatStore.AddChat(chat)
	} else {
		chat := chatStore.Chats[chatID]
		chat.ScrambledWord = scrambleRandomWord
		chatStore.UpdateChat(chatID, chat)
	}

	scrambledWord := s.Join(scrambledCharList, "")
	msg := tgbotapi.NewMessage(chatID, "unscramble: "+scrambledWord)
	bot.Send(msg)
}

func forfeitTabooCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	if !chatStore.ChatFound(update.Message.Chat.ID) {
		chat := models.Chat{
			ID: update.Message.Chat.ID,
		}
		chatStore.AddChat(chat)

	} else {
		chat := chatStore.Chats[chatID]
		chat.TabooWord = ""
		chat.TabooUserID = 0
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("The word that was sent out was %s", chatStore.Chats[chatID].TabooWord))
		bot.Send(msg)
		chatStore.UpdateChat(chatID, chat)
	}
}

func forfeitScrambleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	if !chatStore.ChatFound(update.Message.Chat.ID) {
		chat := models.Chat{
			ID: update.Message.Chat.ID,
		}
		chatStore.AddChat(chat)

	} else {
		chat := chatStore.Chats[chatID]
		chat.ScrambledWord = ""
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("The scrambled word was %s", chatStore.Chats[chatID].ScrambledWord))
		bot.Send(msg)
		chatStore.UpdateChat(chatID, chat)
	}
}

func randomWord() string {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(randomWordArray))
	randomWord := randomWordArray[randomIndex]
	return randomWord
}
