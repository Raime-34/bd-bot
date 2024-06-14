package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	users                  map[int64]*UserInfo
	authorizedMembersGroup map[int64]map[int64]struct{}
	mutex                  sync.RWMutex
)

func main() {
	bot, err := tgbotapi.NewBotAPI("7275504206:AAGQz5i4QFzyXr3MTaof7y_36HQAz3bWW28")
	if err != nil {
		log.Panic(err)
	}
	users = make(map[int64]*UserInfo)
	authorizedMembersGroup = make(map[int64]map[int64]struct{})

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go func() {
		for true {
			trackBD(bot)
		}
	}()

	for update := range updates {

		go func(update tgbotapi.Update) {
			var resMsg tgbotapi.MessageConfig
			mutex.Lock()
			defer mutex.Unlock()

			if upd := update.MyChatMember; upd != nil {
				if upd.NewChatMember.User.IsBot && upd.NewChatMember.Status == "member" {
					resMsg = tgbotapi.NewMessage(upd.Chat.ID, messageAfterJoinChat)
					buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonURL(addBDButton, fmt.Sprintf(newUserLink, upd.Chat.ID))))
					resMsg.ReplyMarkup = buttons
					addMember(upd.Chat.ID, upd.From.ID)
					addUserInfo(&upd.From)
					if _, err := bot.Send(resMsg); err != nil {
						log.Panic(err)
					}
				}
			} else {
				resMsg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
				msg := update.Message

				if msg == nil || msg.From.IsBot {
					return
				}

				if msg.Text == "" {
					return
				}

				addUserInfo(msg.From)

				user := users[msg.From.ID]

				if !msg.IsCommand() {
					switch user.flag {
					case PRINT_YEAR_STAGE:
						year, err := parseInt(msg.Text)
						if err == nil {
							user.birthday = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
							user.flag = PRINT_MONTH_STAGE
							resMsg.Text = enterMonthMessage
							resMsg.ReplyMarkup = monthKeyboard
						} else {
							resMsg.Text = yearParseErrorMessage
						}
					case PRINT_MONTH_STAGE:
						month, err := parseMonth(msg.Text)
						if err != nil {
							resMsg.Text = err.Error()
						} else {
							resMsg.Text = enterDayMessage
							user.flag = PRINT_DAY_STAGE
							user.birthday = time.Date(user.birthday.Year(), month, 1, 0, 0, 0, 0, time.UTC)
							var daysButtons [][]tgbotapi.KeyboardButton
							var buff []tgbotapi.KeyboardButton
							for i := range amountOfDays(user.birthday) {
								buff = append(buff, tgbotapi.NewKeyboardButton(strconv.Itoa(i+1)))
								if len(buff) == 7 {
									daysButtons = append(daysButtons, tgbotapi.NewKeyboardButtonRow(buff...))
									buff = buff[:0]
								}
							}
							if len(buff) != 0 {
								daysButtons = append(daysButtons, tgbotapi.NewKeyboardButtonRow(buff...))
							}
							dayKeyboard := tgbotapi.NewReplyKeyboard(daysButtons...)
							resMsg.ReplyMarkup = dayKeyboard
						}
					case PRINT_DAY_STAGE:
						day, err := parseInt(msg.Text)
						maxAmountOfDays := amountOfDays(user.birthday)
						if err == nil && day <= maxAmountOfDays {
							user.birthday = time.Date(user.birthday.Year(), user.birthday.Month(), day, 0, 0, 0, 0, time.UTC)
							resMsg.Text = doneMessage
							user.flag = DEFAULT_STAGE
							resMsg.ReplyMarkup = hideKeyboard
						} else {
							resMsg.Text = dayParseErrorMessage
						}
					case SUB_STAGE:
						username := strings.Split(msg.Text, " ")[0]
						i, err := findUserByUsername(username)
						if err == nil {
							users[i].subscribers[msg.From.ID] = struct{}{}
							resMsg.Text = fmt.Sprintf(subscriptionDone, users[i].firstName)
							user.flag = DEFAULT_STAGE
							resMsg.ReplyMarkup = hideKeyboard
						} else {
							resMsg.Text = err.Error()
						}
					default:
						resMsg.Text = iDontKnowMessage
					}
				} else {
					switch msg.Command() {
					case "help":
						resMsg.Text = helpMessage
					case "start":
						if args := msg.CommandArguments(); args != "" {
							if id, err := strconv.ParseInt(args, 10, 64); err == nil {
								addMember(id, msg.From.ID)
							}
						}
						resMsg.Text = fmt.Sprintf(greetingMessage, msg.From.FirstName)
					case "setbd":
						resMsg.Text = enterYearMessage
						user.flag = PRINT_YEAR_STAGE
					case "check":
						resMsg.Text = fmt.Sprintf(printBDMessage, user.birthday.Day(), months[user.birthday.Month()], user.birthday.Year())
					case "subscribe":
						user.flag = SUB_STAGE
						resMsg.Text = printUsernameMessage
						availableSubscriptions := findAvailableSubscriptions(user.id)
						if len(availableSubscriptions) > 0 {
							var subButtons [][]tgbotapi.KeyboardButton
							var buff = make([]tgbotapi.KeyboardButton, 0)
							for i := range availableSubscriptions {
								forSub := users[availableSubscriptions[i]]
								if forSub.username != "" {
									buff = append(buff, tgbotapi.NewKeyboardButton(fmt.Sprintf(text4SubButtonWithUsername, forSub.username, forSub.firstName)))
								} else {
									buff = append(buff, tgbotapi.NewKeyboardButton(fmt.Sprintf(text4SubButtonWithoutUsername, forSub.firstName)))
								}
								if len(buff) == 7 {
									subButtons = append(subButtons, tgbotapi.NewKeyboardButtonRow(buff...))
									buff = buff[:0]
								}
							}
							if len(buff) != 0 {
								subButtons = append(subButtons, tgbotapi.NewKeyboardButtonRow(buff...))
							}
							subKeyboard := tgbotapi.NewReplyKeyboard(subButtons...)
							resMsg.ReplyMarkup = subKeyboard
						}
					default:
						resMsg.Text = unknownCommandMessage
					}
				}

				users[msg.From.ID] = user

				if _, err := bot.Send(resMsg); err != nil {
					log.Panic(err)
				}
			}
		}(update)

	}
}
