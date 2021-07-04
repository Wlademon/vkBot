package api

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Wlademon/vkBot/time"
)

const DATE_PERIOD_B = 15

type User interface {
	GetId() string
	GetFullName() string
	GetDateBirth() (string, error)
}

type UserApi interface {
	GetApiName() string
	GetUsers() ([]User, error)
}

var Apis []UserApi

func AddApi(api UserApi) {
	Apis = append(Apis, api)
}

func GetKeys() []string {
	var keys []string
	for _, user := range Apis {
		keys = append(keys, user.GetApiName())
	}

	return keys
}

func GetApiMap() map[string]UserApi {
	var ApiMap = make(map[string]UserApi)
	for _, user := range Apis {
		ApiMap[user.GetApiName()] = user
	}

	return ApiMap
}

func GetApis() []UserApi {
	return Apis
}

func BDateMessage(api UserApi) string {
	users, err := api.GetUsers()
	if err != nil {
		panic(err)
	}

	out := "В ближайшие " + strconv.Itoa(DATE_PERIOD_B) + " дней, дни рождения у:\n"
	sort.Sort(ByDate(users))

	var nowBD []string
	for _, userB := range users {
		date, _ := userB.GetDateBirth()
		tDate, _ := time.Parse("2006-01-02", date)
		if date == time.Now().Format("2006-01-02") {
			date = "Сегодня"
			nowBD = append(nowBD, userB.GetFullName())
		} else {
			date = tDate.Format("02.01.2006") + " числа"
		}
		out += fmt.Sprintf("%s - %s", userB.GetFullName(), date) + "\n"
	}
	if len(nowBD) > 0 {
		out += "\n" + addMessage(nowBD)
	}

	return out
}

func addMessage(people []string) string {
	if len(people) > 0 {
		lastNow := people[len(people)-1]
		obMess := "Мы рады, что ты с нами работаешь. "
		if len(people) > 1 {
			obMess = "Мы рады, что вы с нами работаете. "
			lastNow = strings.Join(people[:len(people)-1], ", ") + " и " + lastNow
		}
		return "С днем рождения: " + lastNow + "!\n" + obMess + "Желаем расти и развиваться."
	}

	return ""
}

type ByDate []User

func (u ByDate) Len() int {
	return len(u)
}

func (u ByDate) Swap(i, j int) { u[i], u[j] = u[j], u[i] }

func (u ByDate) Less(i, j int) bool {
	f, _ := u[i].GetDateBirth()
	s, _ := u[j].GetDateBirth()

	return f < s
}
