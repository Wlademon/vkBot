package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const BITRIX_API = "bitrix"

// User

type BitrixUser struct {
	Id         int64  `json:"ID"`
	FirstName  string `json:"NAME"`
	LastName   string `json:"LAST_NAME"`
	MiddleName string `json:"SECOND_NAME"`
	BDate      string `json:"PERSONAL_BIRTHDAY"`
}

func (u BitrixUser) GetId() string {
	return strconv.FormatInt(u.Id, 10)
}

func (u BitrixUser) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

func (u BitrixUser) GetDateBirth() (string, error) {
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

// Interlayer

type BitrixReq struct {
	Result []BitrixUser `json:"result"`
}

// API User

type Bitrix struct {
	Url string
}

func (b Bitrix) GetApiName() string {
	return BITRIX_API
}

func (b Bitrix) getUrl() string {
	return b.Url
}

func (b Bitrix) GetUsers() ([]User, error) {
	resp, err := http.Get(b.getUrl())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	user := new(BitrixReq)
	var userBA []User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	for _, u := range user.Result {
		date, err := u.GetDateBirth()
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
