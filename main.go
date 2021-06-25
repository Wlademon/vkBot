package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/go-vk-api/vk"
	"github.com/joho/godotenv"
)

const BITRIX_API = "bitrix"
const VK_API = "vk"
const DATE_PERIOD_B = 15

var VkClient *vk.Client
var UserData User
var ChatId = make(map[string][]int)

type User interface {
	getFriendsBirthDate() ([]BUser, error)
	getType() string
}

type BitrixUser struct {
	Id         int64  `json:"ID"`
	FirstName  string `json:"NAME"`
	LastName   string `json:"LAST_NAME"`
	MiddleName string `json:"SECOND_NAME"`
	BDate      string `json:"PERSONAL_BIRTHDAY"`
}

func (u BitrixUser) getId() int64 {
	return u.Id
}

func (u BitrixUser) getFullName() string {
	return u.FirstName + " " + u.LastName
}

func (u BitrixUser) getDate() (string, error) {
	if len(u.BDate) == 0 {
		return "", errors.New("date not found")
	}
	timeB, err := time.Parse("2006-01-02T15:04:05", strings.Split(u.BDate, "+")[0])
	if err != nil {
		return "", err
	}
	dateArr := strings.Split(timeB.Format("2006-01-02"), "-")
	dateArr[0] = strconv.Itoa(time.Now().Year())

	return strings.Join(dateArr, "-"), nil
}

type BitrixReq struct {
	Result []BitrixUser `json:"result"`
}

type Bitrix struct {
	Url string
}

func (b Bitrix) getUrl() string {
	return b.Url
}

func (b Bitrix) getType() string {
	return BITRIX_API
}

func (b Bitrix) getFriendsBirthDate() ([]BUser, error) {
	fmt.Println(b.getUrl())
	resp, err := http.Get(b.getUrl())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	user := new(BitrixReq)
	var userBA []BUser
	json.NewDecoder(resp.Body).Decode(&user)
	for _, u := range user.Result {
		date, err := u.getDate()
		if err != nil {
			continue
		}
		DTime, err := time.Parse("2006-01-02", date)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if DTime.Format("2006-01-02") >= time.Now().Format("2006-01-02") &&
			DTime.Format("2006-01-02") <= time.Now().AddDate(0, 0, DATE_PERIOD_B).Format("2006-01-02") {
			fmt.Println(DTime.Format("2006-01-02"))
			userBA = append(userBA, u)
		}
	}

	return userBA, nil
}

type BUser interface {
	getId() int64
	getFullName() string
	getDate() (string, error)
}

type ResponseItems struct {
	Items []Friend `json:"items"`
}

type Friend struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	DateBerth string `json:"bdate"`
}

func (f Friend) getId() int64 {
	return f.Id
}

func (f Friend) getFullName() string {
	return f.FirstName + " " + f.LastName
}

func (f Friend) getDate() (string, error) {
	if len(f.DateBerth) == 0 {
		return "", errors.New("date not found")
	}
	dmy := strings.Split(f.DateBerth, ".")
	if len(dmy) == 2 {
		dmy = append(dmy, time.Now().Format("2006"))
	} else if len(dmy) == 3 {
		dmy[2] = time.Now().Format("2006")
	} else {
		return "", errors.New("date in non format")
	}
	if len(dmy[0]) == 1 {
		dmy[0] = "0" + dmy[0]
	}
	if len(dmy[1]) == 1 {
		dmy[1] = "0" + dmy[1]
	}

	return strings.Join(dmy, "-"), nil
}

type CurrentUser struct {
	Id int64 `json:"id"`
}

func (u CurrentUser) getId() int64 {
	return u.Id
}

func initEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	ChatId[BITRIX_API] = []int{}
	ChatId[VK_API] = []int{}
}

func initClients() *vk.Client {
	client, _ := vk.NewClientWithOptions(
		vk.WithToken(os.Getenv("VK_ACCESS_TOKEN")),
	)
	return client
}

func initBot() (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_ACCESS_TOKEN"))
	if err != nil {
		return nil, err
	}
	bot.Debug = true

	return bot, err
}

func (u CurrentUser) getType() string {
	return VK_API
}

func (u CurrentUser) getFriendsBirthDate() ([]BUser, error) {
	var Friends []Friend
	var resItems ResponseItems
	var BDateFriend []BUser
	params := vk.RequestParams{
		"count":     1000,
		"fields":    "bdate",
		"name_case": "nom",
		"v":         "5.130",
		"order":     "name",
	}
	err := VkClient.CallMethod("friends.get", params, &resItems)
	if err != nil {
		return nil, err
	}

	Friends = resItems.Items

	for _, user := range Friends {
		date, err := user.getDate()
		if err != nil {
			continue
		}
		DTime, err := time.Parse("02-01-2006", date)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(date)
		if DTime.Format("2006-01-02") >= time.Now().Format("2006-01-02") &&
			DTime.Format("2006-01-02") <= time.Now().AddDate(0, 0, 15).Format("2006-01-02") {
			BDateFriend = append(BDateFriend, user)
		}
	}

	return BDateFriend, nil
}

func getUserData() (User, error) {
	var err error
	var user User
	var users []CurrentUser
	params := vk.RequestParams{
		"fields": "photo_50,verified,about,bdate",
	}
	err = VkClient.CallMethod("users.get", params, &users)
	if err != nil {
		return nil, err
	} else if len(users) == 0 {
		return nil, errors.New("zero users")
	}

	user = users[0]

	return user, err
}

func dRemember(bot *tgbotapi.BotAPI, users []User) {
	var messageVk string
	var messageBitrix string
	time.Sleep(60 * time.Second)
	for true {
		messageVk = ""
		messageBitrix = ""
		for _, user := range users {
			if user.getType() == BITRIX_API {
				messageBitrix += bdate(user)
			} else if user.getType() == VK_API {
				messageVk += bdate(user)
			}
		}
		for key, ids := range ChatId {
			switch key {
			case VK_API:
				for _, id := range ids {
					if id != 0 && len(messageVk) != 0 {
						msg := tgbotapi.NewMessage(int64(id), "VK: \n"+messageVk)
						_, _ = bot.Send(msg)
					}
				}
				break
			case BITRIX_API:
				for _, id := range ids {
					if id != 0 && len(messageBitrix) != 0 {

						msg := tgbotapi.NewMessage(int64(id), "Bitrix: \n"+messageBitrix)
						_, _ = bot.Send(msg)
					}
				}

			}
		}

		time.Sleep(24 * time.Hour)
	}
}

func main() {
	initEnv()
	bot, err := initBot()
	if err != nil {
		panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	VkClient = initClients()
	user, err := getUserData()
	UserData = user
	var Users []User
	bitrix := Bitrix{Url: os.Getenv("BITRIX_URL")}
	metaUser := MetaUser{
		vk:     user,
		bitrix: bitrix,
	}
	Users = append(Users, user)
	Users = append(Users, bitrix)
	go dRemember(bot, Users)
	var isCommand bool
	for update := range updates {
		if update.Message == nil {
			continue
		}
		fmt.Println(update.Message.Entities)
		entities := update.Message.Entities
		isCommand = false
		if entities != nil {
			for _, entity := range *entities {
				if entity.Type == "bot_command" {
					isCommand = true
					break
				}
			}
		}
		if isCommand {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, runCommand(update.Message.Text, update.Message.Chat, metaUser))
			msg.ReplyToMessageID = update.Message.MessageID
			_, _ = bot.Send(msg)
		}
	}
}

type ByDate []BUser

func (u ByDate) Len() int {
	return len(u)
}

func (u ByDate) Swap(i, j int) { u[i], u[j] = u[j], u[i] }

func (u ByDate) Less(i, j int) bool {
	f, _ := u[i].getDate()
	s, _ := u[j].getDate()

	return f < s
}

func bdate(UserData User) string {
	users, err := UserData.getFriendsBirthDate()
	if err != nil {
		panic(err)
	}
	out := "В ближайшие " + strconv.Itoa(DATE_PERIOD_B) + " дней, дни рождения у:\n"
	sort.Sort(ByDate(users))
	var nowBD []string
	for _, userB := range users {
		date, _ := userB.getDate()
		tDate, _ := time.Parse("2006-01-02", date)
		if date == time.Now().Format("2006-01-02") {
			date = "Сегодня"
			nowBD = append(nowBD, userB.getFullName())
		} else {
			date = tDate.Format("02.01.2006") + " числа"
		}
		out += fmt.Sprintf("%s - %s", userB.getFullName(), date) + "\n"
	}
	if len(nowBD) > 0 {
		lastNow := nowBD[len(nowBD)-1]
		obMess := " Мы рады, что ты с нами работаешь. "
		if len(nowBD) > 1 {
			obMess = " Мы рады, что вы с нами работаете. "
			lastNow = strings.Join(nowBD[:len(nowBD)-1], ", ") + " и " + lastNow
		}
		out += "\n С днем рождения \n" + lastNow + "!\n" + obMess + "Желаем расти и развиваться."
	}

	return out
}

type MetaUser struct {
	vk     User
	bitrix Bitrix
}

func runCommand(command string, chat *tgbotapi.Chat, user MetaUser) string {
	arrCommand := strings.Split(strings.Trim(strings.ReplaceAll(command, "  ", " "), " "), " ")
	commandExec := arrCommand[0]
	commandExec = strings.Split(commandExec, "@")[0]
	var args []string
	if len(arrCommand) > 1 {
		args = arrCommand[1:]
	}
	switch commandExec {
	case "/kill":
		os.Exit(9)
	case "/test":
		return "Иди нах с такими тестами"
	case "/date":
		return bdate(user.bitrix)
	case "/bdate":
		return bdate(user.vk)
	case "/cview":
		return strconv.Itoa(int(chat.ID))
	case "/save":
		for _, v := range args {
			if inArrayString([]string{BITRIX_API, VK_API}, v) != -1 {
				if inArray(ChatId[v], int(chat.ID)) == -1 {
					iter := inArray(ChatId[v], int(chat.ID))
					if inArray(ChatId[v], iter) != -1 {
						ChatId[v][iter] = int(chat.ID)
					} else {
						ChatId[v] = append(ChatId[v], int(chat.ID))
					}
				}
			}
		}
		return "Сохранено"
	case "/view":
		result := ""
		for i, v := range ChatId {
			if inArray(v, int(chat.ID)) != -1 {
				result += fmt.Sprintf("%s - %s \n", i, strconv.Itoa(int(chat.ID)))
			}
		}
		if len(result) == 0 {
			result = "список пуст"
		}

		return result
	case "/clean":
		result := ""
		for i, v := range ChatId {
			iter := inArray(v, int(chat.ID))
			if len(args) == 0 {
				if iter != -1 {
					v[iter] = 0
					ChatId[i] = v
				}
				result = "Вы полностью удалены"
			} else {
				for _, key := range args {
					if i == key {
						if iter != -1 {
							v[iter] = 0
							ChatId[i] = v
						}
						result += "Вы удалены из " + i + "\n"
					}
				}
			}
		}

		return result

	}

	return "команда - фигня"
}

func inArray(num []int, an int) int {
	exist := false
	iter := 0
	for i, n := range num {
		if n == an {
			iter = i
			exist = true
			break
		}
	}
	if exist {
		return iter
	}
	return -1
}

func inArrayString(num []string, an string) int {
	exist := false
	iter := 0
	for i, n := range num {
		if n == an {
			iter = i
			exist = true
			break
		}
	}
	if exist {
		return iter
	}
	return -1
}
