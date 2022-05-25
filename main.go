package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const statsCommand string = "/stats"

var lenStatsCommand int = len(statsCommand)

const startCommand string = "/start"

var lenStartCommand int = len(startCommand)

const botTag string = "@CryptoYVBot"

var lenBotTag int = len(botTag)

const telegramApiBaseUrl string = "https://api.telegram.org/bot"
const telegramApiSendMessage string = "/sendMessage"
const telegramTokenEnv string = "TOKEN"

var telegramApi string = telegramApiBaseUrl + os.Getenv(telegramTokenEnv) + telegramApiSendMessage

var cryptoStatsApi string = "http://localhost:8080/cryptostats?symbol="

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}

type Chat struct {
	Id int `json:"id"`
}

type Response struct {
	Status            string `json:"status"`
	Message           string `json:"message"`
	CurrencyName      string `json:"currencyName"`
	Price             string `json:"price"`
	Change            string `json:"change"`
	ChangePercentage  string `json:"changePercentage"`
	PrevClose         string `json:"prevClose"`
	Open              string `json:"open"`
	DayRange          string `json:"dayRange"`
	YearRange         string `json:"yearRange"`
	StartDate         string `json:"startDate"`
	MarketCap         string `json:"marketCap"`
	CirculatingSupply string `json:"circulatingSupply"`
	Volume            string `json:"volume"`
	Desc              string `json:"desc"`
}

func parseRequest(r *http.Request) (*Update, error) {
	var update Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("could not decode incoming update %s", err.Error())
		return nil, err
	}
	return &update, nil
}

func handleWebHook(w http.ResponseWriter, r *http.Request) {
	var update, err = parseRequest(r)
	if err != nil {
		log.Printf("error in parsing update %s", err.Error())
		return
	}
	var symbol = clean(update.Message.Text)
	var stats, errText = getStats(symbol)
	if errText != nil {
		log.Printf("error in calling crypto-stats-api")
		return
	}
	// sending stats back
	var telegramResponse, errTelegram = sendText(update.Message.Chat.Id, stats)
	if errTelegram != nil {
		log.Printf("encountered error %s, response body is %s", errTelegram.Error(), telegramResponse)
	} else {
		log.Printf("stats %s successfully distributed to chat id %d", stats, update.Message.Chat.Id)
	}

}

func getStats(symbol string) (string, error) {
	resp, err := http.Get(cryptoStatsApi + symbol)
	if err != nil {
		fmt.Println("Could not hit endpoint")
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func sendText(chatId int, text string) (string, error) {
	log.Printf("sending %s to chad_id: %d", text, chatId)
	response, err := http.PostForm(
		telegramApi,
		url.Values{
			"chat_id": {strconv.Itoa(chatId)},
			"text":    {text},
		})
	if err != nil {
		log.Printf("error in sending text to chat, %s", err.Error())
		return "", err
	}
	defer response.Body.Close()
	var bodyBytes, errRead = ioutil.ReadAll(response.Body)
	if errRead != nil {
		log.Printf("error in parsing answer, %s", errRead.Error())
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Printf("body of response is: %s", bodyString)
	return bodyString, nil
}

func clean(s string) string {
	if len(s) >= lenStartCommand {
		if s[:lenStartCommand] == startCommand {
			s = s[lenStartCommand:]
		}
	}
	if len(s) >= lenStatsCommand {
		if s[:lenStatsCommand] == statsCommand {
			s = s[lenStatsCommand:]
		}
	}
	if len(s) >= lenBotTag {
		if s[:lenBotTag] == botTag {
			s = s[lenBotTag:]
		}
	}
	return s + "-USD"
}
