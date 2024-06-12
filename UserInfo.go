package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"slices"
	"time"
)

type UserInfo struct {
	id          int64
	firstName   string
	birthday    time.Time
	flag        int
	username    string
	subscribers map[int64]struct{}
}

func newUser(user *tgbotapi.User) *UserInfo {
	return &UserInfo{
		id:          user.ID,
		firstName:   user.FirstName,
		username:    user.UserName,
		subscribers: make(map[int64]struct{}),
	}
}

func addMember(chat int64, member int64) {
	_, ok := authorizedMembersGroup[chat]
	if !ok {
		authorizedMembersGroup[chat] = make(map[int64]struct{})
	}
	authorizedMembersGroup[chat][member] = struct{}{}
}

func findAvailableSubscriptions(member int64) (result []int64) {
	result = make([]int64, 0)
	for _, members := range authorizedMembersGroup {
		var keys = make([]int64, 0)
		for i, _ := range members {
			keys = append(keys, i)
		}
		slices.Sort(keys)
		if n, ok := slices.BinarySearch(keys, member); ok {
			keys = append(keys[:n], keys[n+1:]...)
			result = append(result, keys...)
		}
	}
	return
}

func addUserInfo(user *tgbotapi.User) {
	if users[user.ID] == nil {
		users[user.ID] = newUser(user)
	}
}
