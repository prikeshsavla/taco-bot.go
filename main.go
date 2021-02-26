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

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var tacoTrades []string

func slackBot(port string) {
	token := os.Getenv("SLACKTOKEN")
	var api = slack.New(token)

	signingSecret := os.Getenv("SIGNING_SECRET")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("failed")
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("failed")
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println("failed")
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
			fmt.Println("challenged")
		}
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("failed")
			return
		}
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("failed")
			return
		}
		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("failed")
			return
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
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
	})
	fmt.Println("[INFO] Server listening")
	fmt.Println(port)
	http.ListenAndServe(":"+port, nil)
}

func sendTaco(api *slack.Client, userFrom string, userTo string) {
	from, err := api.GetUserInfo(userFrom)

	tacoTrades = append(tacoTrades, userFrom)

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
	slackBot(port)

}
