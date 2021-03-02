package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Trade struct {
	from string
	to   string
	at   time.Time
}

var tacoTrades []Trade

func slackBot(port string) {
	token := os.Getenv("SLACKTOKEN")
	var api = slack.New(token)

	signingSecret := os.Getenv("SIGNING_SECRET")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest)
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			writeError(w, http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				writeError(w, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
			fmt.Println("challenged")
		}
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			writeError(w, http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			writeError(w, http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			writeError(w, http.StatusUnauthorized)
			return
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			receiveTaco(api, eventsAPIEvent)
		}
	})
	fmt.Println("[INFO] Server listening")
	fmt.Println(port)
	fmt.Println("Go Redis Tutorial")

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	// we can call set with a `Key` and a `Value`.
	err = client.Set("name", "Elliot", 0).Err()
	// if there has been an error setting the value
	// handle the error
	if err != nil {
		fmt.Println(err)
	}

	val, err := client.Get("name").Result()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(val)

	http.ListenAndServe(":"+port, nil)
}

func writeError(w http.ResponseWriter, errorCode int) {
	w.WriteHeader(errorCode)
	fmt.Println("failed")
}

func receiveTaco(api *slack.Client, eventsAPIEvent slackevents.EventsAPIEvent) {

	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		{
			api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
		}
	case *slackevents.MessageEvent:
		{
			fmt.Println("Message", ev.Text)
			fmt.Println("Message User", ev.User)

			if strings.Contains(ev.Text, ":taco:") {
				api.PostMessage(ev.User, slack.MsgOptionText("Yes, Here is a taco for ", false))
			}

		}
	case *slackevents.ReactionAddedEvent:
		{
			fmt.Println("Reaction", ev.Reaction)
			if strings.Contains(ev.Reaction, "white_check_mark") {

				sendTaco(api, ev.User, ev.ItemUser)
			}
		}
	}

}

func sendTaco(api *slack.Client, userFrom string, userTo string) {
	from, err := api.GetUserInfo(userFrom)

	tacoTrades = append(tacoTrades, Trade{from: userFrom})

	fmt.Println(tacoTrades)
	if err != nil {
		log.Println(err)
	}
	to, err := api.GetUserInfo(userTo)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(from.RealName)
	fmt.Println(to.RealName)
	api.PostMessage(userFrom, slack.MsgOptionText("You sent a taco to  <@"+userFrom+">, "+strconv.Itoa(5-len(tacoTrades))+" Tacos Left", false))
	api.PostMessage(userTo, slack.MsgOptionText("<@"+userFrom+"> gave you a taco", false))
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	// slackBot(port)
	http.HandleFunc("/", dummyBot)

	http.ListenAndServe(":"+port, nil)
}

func dummyBot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var query = r.URL.Query()

	fmt.Println("from: " + query.Get("from"))
	fmt.Println("to: " + query.Get("to"))

	tacoTrades = append(tacoTrades, Trade{from: query.Get("from"), to: query.Get("to"), at: time.Now()})

	fmt.Println(tacoTrades)

	w.Write([]byte("Received a GET request\n"))

}
