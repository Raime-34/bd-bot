package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	monthKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(months[time.January]),
			tgbotapi.NewKeyboardButton(months[time.February]),
			tgbotapi.NewKeyboardButton(months[time.March]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(months[time.April]),
			tgbotapi.NewKeyboardButton(months[time.May]),
			tgbotapi.NewKeyboardButton(months[time.June]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(months[time.July]),
			tgbotapi.NewKeyboardButton(months[time.August]),
			tgbotapi.NewKeyboardButton(months[time.September]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(months[time.October]),
			tgbotapi.NewKeyboardButton(months[time.November]),
			tgbotapi.NewKeyboardButton(months[time.December]),
		),
	)
	hideKeyboard = tgbotapi.NewRemoveKeyboard(true)
)

func parseInt(msg string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(msg))
	if err != nil {
		return -1, err
	} else {
		return n, nil
	}
}

func parseMonth(msg string) (time.Month, error) {
	switch strings.ToLower(msg) {
	case "январь":
		return time.January, nil
	case "февраль":
		return time.February, nil
	case "март":
		return time.March, nil
	case "апрель":
		return time.April, nil
	case "май":
		return time.May, nil
	case "июнь":
		return time.June, nil
	case "июль":
		return time.July, nil
	case "август":
		return time.August, nil
	case "сентябрь":
		return time.September, nil
	case "октябрь":
		return time.October, nil
	case "ноябрь":
		return time.November, nil
	case "декабрь":
		return time.December, nil
	default:
		return 0, fmt.Errorf(monthParseErrorMessage, msg)
	}
}

func amountOfDays(date time.Time) int {
	nextMonth := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	nextMonth = nextMonth.AddDate(0, 0, -1)
	return nextMonth.Day()
}

func findUserByUsername(username string) (int64, error) {
	for i, info := range users {
		if info.username == username || info.username == "" && info.firstName == username {
			return i, nil
		}
	}
	return -1, fmt.Errorf(userNotFoundMessage)
}

func trackBD(bot *tgbotapi.BotAPI) {
	<-time.After(5 * time.Minute)
	mutex.RLock()
	defer mutex.RUnlock()
	for _, info := range users {
		if info.birthday.Day() != time.Now().Day() || info.birthday.Month() != time.Now().Month() {
			continue
		}
		if len(info.subscribers) > 0 {
			for id, _ := range info.subscribers {
				if _, err := bot.Send(tgbotapi.NewMessage(id, fmt.Sprintf(notificationMessage, info.username))); err != nil {
					log.Panic(err)
				}
			}
		}
	}
}
