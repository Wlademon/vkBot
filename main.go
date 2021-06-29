package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"

	"vkBot/api"
	"vkBot/file/cache"
)

const defHour = 8
const cacheHourPrefix = "_cacheHour_"

var ChatId = make(map[string][]int)
var ChatTime = make(map[string][2]int)

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
				messages[user.GetApiName()] = api.BDateMessage(user)
				err := cache.Create("MESSAGES_"+user.GetApiName(), messages[user.GetApiName()], time.Hour).Set()
				if err != nil {
					fmt.Println(err)
				}
			} else {
				messages[user.GetApiName()] = message
			}
		}
		for key, ids := range ChatId {
			if messages[key] == "" {
				continue
			}
			var message = messages[key]
			for _, id := range ids {
				if exist, _ := cache.Get(cacheHourPrefix + strconv.Itoa(id)); !exist {
					sId := strconv.Itoa(id)
					hour := defHour
					min := 0
					if ChatTime[sId] != [2]int{} {
						HM := ChatTime[sId]
						hour = HM[0]
						min = HM[1]
					}
					if time.Now().Hour() >= hour && time.Now().Minute() >= min {
						if id != 0 && len(message) != 0 {
							fmt.Println("MESSAGE SEND")
							msg := tgbotapi.NewMessage(int64(id), key+": \n"+message)
							_, _ = bot.Send(msg)
						}
						err := cache.Create(cacheHourPrefix+strconv.Itoa(id), "sended", 23*time.Hour+59*time.Minute).Set()
						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
		}
		fmt.Println("FINISH")
		time.Sleep(time.Second * 60)
	}
}

func main() {
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
			returnMessage := runCommand(update.Message.Text, update.Message.Chat, api.GetApiMap())
			if returnMessage != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, returnMessage)
				msg.ReplyToMessageID = update.Message.MessageID
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
		return
	}
	err = cache.CreateForever("CHAT_TIME", string(marshal)).Set()
	if err != nil {
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

func runCommand(command string, chat *tgbotapi.Chat, user map[string]api.UserApi) string {
	arrCommand := strings.Split(strings.Trim(strings.ReplaceAll(command, "  ", " "), " "), " ")
	commandExec := arrCommand[0]
	commandExec = strings.Split(commandExec, "@")[0]
	var args []string
	if len(arrCommand) > 1 {
		args = arrCommand[1:]
	}
	switch commandExec {
	case "/hour":
		if len(args) == 0 {
			return ""
		}
		if hour, err := strconv.Atoi(args[0]); err == nil {
			var HM [2]int
			HM = [2]int{defHour, 0}
			if hour < 24 && hour >= 0 {
				HM = [2]int{hour, 0}
			}
			if minute, errM := strconv.Atoi(args[1]); errM == nil && minute > 0 && minute < 60 {
				HM[1] = minute
			}
			ChatTime[strconv.FormatInt(chat.ID, 10)] = HM
			cache.Flush(cacheHourPrefix + strconv.FormatInt(chat.ID, 10))
			setChatTimeCache()
			return "Время задано"
		}
	case "/kill":
		setChatIdCache()
		setChatTimeCache()
		os.Exit(9)
	case "/test":
		return "Иди нах с такими тестами"
	case "/date":
		return api.BDateMessage(user[api.BITRIX_API])
	case "/bdate":
		return api.BDateMessage(user[api.VK_API])
	case "/cview":
		return strconv.Itoa(int(chat.ID))
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
			setChatIdCache()
		}

		return result

	}

	return ""
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
