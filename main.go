package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	defTime "time"

	"github.com/Wlademon/vkBot/time"

	"github.com/Wlademon/vkBot/api"
	"github.com/Wlademon/vkBot/file/cache"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

const defHour = 8
const cacheHourPrefix = "_cacheHour_"

var ChatId = make(map[string][]int)
var ChatTime = make(map[string][3]int)

func initEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
}

func initBot() (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_ACCESS_TOKEN"))
	if err != nil {
		return nil, err
	}
	bot.Debug = true

	return bot, err
}

func dRemember(bot *tgbotapi.BotAPI, users []api.UserApi) {
	for true {
		fmt.Println("START")
		var messages = make(map[string]string)
		for _, user := range users {
			if existMessage, message := cache.Get("MESSAGES_" + user.GetApiName()); !existMessage {
				fmt.Println("MESSAGE GET")
				messages[user.GetApiName()] = api.BDateMessage(user)
				err := cache.Create("MESSAGES_"+user.GetApiName(), messages[user.GetApiName()], defTime.Hour).Set()
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("MESSAGE exist")
				messages[user.GetApiName()] = message
			}
		}
		fmt.Println("Chats:", ChatId)
		for key, ids := range ChatId {
			fmt.Println(key, messages[key])
			if messages[key] == "" {
				continue
			}
			var message = messages[key]
			fmt.Println(ids)
			for _, id := range ids {
				if exist, _ := cache.Get(cacheHourPrefix + strconv.Itoa(id)); !exist {
					fmt.Println("CACHE not exist")
					sId := strconv.Itoa(id)
					hour := defHour
					min := 0
					if ChatTime[sId][2] != 0 {
						HM := ChatTime[sId]
						hour = HM[0]
						min = HM[1]
					}
					fmt.Println(time.Now().Format("15:04:05"))
					fmt.Println(hour, min)
					if time.Now().Hour() >= hour && time.Now().Minute() >= min {
						if id != 0 && len(message) != 0 {
							fmt.Println("MESSAGE SEND")
							msg := tgbotapi.NewMessage(int64(id), key+": \n"+message)
							_, _ = bot.Send(msg)
						}
						fmt.Println("CREATE CACHE:")
						err := cache.Create(cacheHourPrefix+strconv.Itoa(id), "sended", 23*defTime.Hour+59*defTime.Minute).Set()
						if err != nil {
							fmt.Println(err)
						}
					}
				} else {
					fmt.Println("CACHE  exist")
				}
			}
		}
		fmt.Println("FINISH")
		defTime.Sleep(defTime.Second * 60)
	}
}

func main() {
	time.InitTime(defTime.Hour * 3 / defTime.Second)
	initEnv()
	cache.InitCache("cache")
	getCache()
	bot, err := initBot()
	if err != nil {
		panic(err)
	}

	// createApi(api.VkUser{Token: os.Getenv("VK_ACCESS_TOKEN")})
	createApi(api.Bitrix{Url: os.Getenv("BITRIX_URL")})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	go dRemember(bot, api.GetApis())

	observeCommands(updates, bot)
}

func observeCommands(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI) {
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
			returnMessage, reply := runCommand(update.Message.Text, update.Message, api.GetApiMap())
			if returnMessage != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, returnMessage)
				if reply {
					msg.ReplyToMessageID = update.Message.MessageID
				}
				_, _ = bot.Send(msg)
			}
		}
	}
}

func setChatIdCache() {
	marshal, err := json.Marshal(ChatId)
	if err != nil {
		return
	}
	err = cache.CreateForever("CHAT_ID", string(marshal)).Set()
	if err != nil {
		return
	}
}

func setChatTimeCache() {
	marshal, err := json.Marshal(ChatTime)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cache.CreateForever("CHAT_TIME", string(marshal)).Set()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getCache() {
	if existId, ids := cache.Get("CHAT_ID"); existId {
		err := json.Unmarshal([]byte(ids), &ChatId)
		if err != nil {
			fmt.Println(err)
		}
	}
	if existTimes, times := cache.Get("CHAT_TIME"); existTimes {
		err := json.Unmarshal([]byte(times), &ChatTime)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func createApi(userApi api.UserApi) {
	api.AddApi(userApi)
	if len(ChatId[userApi.GetApiName()]) == 0 {
		ChatId[userApi.GetApiName()] = []int{}
	}
}

func runCommand(command string, message *tgbotapi.Message, user map[string]api.UserApi) (string, bool) {
	chat := message.Chat
	arrCommand := strings.Split(strings.Trim(strings.ReplaceAll(command, "  ", " "), " "), " ")
	commandExec := arrCommand[0]
	commandExec = strings.Split(commandExec, "@")[0]
	var args []string
	if len(arrCommand) > 1 {
		args = arrCommand[1:]
	}
	switch commandExec {
	case "/fuckoff":
		Entities := message.Entities

		var users []string
		for _, entry := range *Entities {
			if entry.Type == "mention" {
				runeText := []rune(message.Text)
				users = append(users, string(runeText[entry.Offset:entry.Offset+entry.Length]))
			}
		}
		if len(users) > 0 {
			if len(users) == 1 {
				return users[0] + " ?????????? ???? ??????.", false
			}

			return strings.Join(users, " , ") + " ?????????? ???? ?????? ??????????.", false
		}

		return "???? ?????? ???????????? ???????? ???????? ?????????? ??????????...", true
	case "/s_done":
		exist, _ := cache.Get(cacheHourPrefix + strconv.FormatInt(chat.ID, 10))
		if exist {
			return "??????????????????", true
		}
		return "?????????????? ????????????????????", true
	case "/schedule":
		temp := ChatTime[strconv.FormatInt(chat.ID, 10)]
		if temp[2] == 0 {
			return "???? ????????????, ?????????? ?????????????????? ?????????????????? ???? ??????????????????", true
		}
		return fmt.Sprintf("hour: %d minute: %d", temp[0], temp[1]), true
	case "/now":
		return time.Now().Format("2006-01-02 15:04:05"), true
	case "/hour":
		if len(args) == 0 {
			return "", false
		}
		if hour, err := strconv.Atoi(args[0]); err == nil {
			var HM [3]int
			HM = [3]int{defHour, 0, 1}
			if hour < 24 && hour >= 0 {
				HM = [3]int{hour, 0, 1}
			}
			if len(args) > 1 {
				if minute, errM := strconv.Atoi(args[1]); errM == nil && minute > 0 && minute < 60 {
					HM[1] = minute
				}
			}
			ChatTime[strconv.FormatInt(chat.ID, 10)] = HM
			cache.Flush(cacheHourPrefix + strconv.FormatInt(chat.ID, 10))
			setChatTimeCache()
			return "?????????? ????????????", true
		}
	case "/kill":
		setChatIdCache()
		setChatTimeCache()
		os.Exit(9)
	case "/test":
		return "?????? ?????? ?? ???????????? ??????????????", true
	case "/date":
		return api.BDateMessage(user[api.BITRIX_API]), true
	case "/bdate":
		return api.BDateMessage(user[api.VK_API]), true
	case "/cview":
		return strconv.Itoa(int(chat.ID)), true
	case "/save":
		for _, v := range args {
			if inArrayString(api.GetKeys(), v) != -1 {
				if inArray(ChatId[v], int(chat.ID)) == -1 {
					iter := inArray(ChatId[v], int(chat.ID))
					if inArray(ChatId[v], iter) != -1 {
						ChatId[v][iter] = int(chat.ID)
					} else {
						ChatId[v] = append(ChatId[v], int(chat.ID))
					}
					setChatIdCache()
				}
			}
		}
		return "??????????????????", true
	case "/view":
		result := ""
		for i, v := range ChatId {
			if inArray(v, int(chat.ID)) != -1 {
				result += fmt.Sprintf("%s - %s \n", i, strconv.Itoa(int(chat.ID)))
			}
		}
		if len(result) == 0 {
			result = "???????????? ????????"
		}

		return result, true
	case "/clean":
		result := ""
		for i, v := range ChatId {
			iter := inArray(v, int(chat.ID))
			if len(args) == 0 {
				if iter != -1 {
					v[iter] = 0
					ChatId[i] = v
				}
				result = "???? ?????????????????? ??????????????"
			} else {
				for _, key := range args {
					if i == key {
						if iter != -1 {
							v[iter] = 0
							ChatId[i] = v
						}
						result += "???? ?????????????? ???? " + i + "\n"
					}
				}
			}
			setChatIdCache()
		}

		return result, true

	}

	return "", false
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
