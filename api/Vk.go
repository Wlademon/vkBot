package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-vk-api/vk"
)

const VK_API = "vk"

var client *vk.Client = nil

// User

type Friend struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	DateBerth string `json:"bdate"`
}

func (f Friend) GetId() string {
	return strconv.FormatInt(f.Id, 10)
}

func (f Friend) GetFullName() string {
	return f.FirstName + " " + f.LastName
}

func (f Friend) GetDateBirth() (string, error) {
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

// Interlayer

type ResponseItems struct {
	Items []Friend `json:"items"`
}

// API user

type VkUser struct {
	Token string
}

func (v VkUser) GetApiName() string {
	return VK_API
}

func (v VkUser) GetUsers() ([]User, error) {
	if client == nil {
		err := v.initClient()
		if err != nil {
			return nil, err
		}
	}
	var Friends []Friend
	var resItems ResponseItems
	var BDateFriend []User
	params := vk.RequestParams{
		"count":     1000,
		"fields":    "bdate",
		"name_case": "nom",
		"v":         "5.130",
		"order":     "name",
	}
	err := client.CallMethod("friends.get", params, &resItems)
	if err != nil {
		return nil, err
	}

	Friends = resItems.Items

	for _, user := range Friends {
		date, err := user.GetDateBirth()
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
			DTime.Format("2006-01-02") <= time.Now().AddDate(0, 0, DATE_PERIOD_B).Format("2006-01-02") {
			BDateFriend = append(BDateFriend, user)
		}
	}

	return BDateFriend, nil

}

func (v VkUser) initClient() error {
	clientVk, err := vk.NewClientWithOptions(
		vk.WithToken(v.Token),
	)

	client = clientVk

	return err
}
