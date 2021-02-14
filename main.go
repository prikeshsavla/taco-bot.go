package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// You more than likely want your "Bot User OAuth Access Token" which starts with "xoxb-"
var api = slack.New("xoxb-59268893569-1719504867733-4kFL22NM2iGqFSSldf13khxZ")

func slackBot(port string) {
	signingSecret := "44d0ac27e7e8b90214e65526a03d83f8"

	http.HandleFunc("/events-endpoint", func(w http.ResponseWriter, r *http.Request) {
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
						api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, Here is a taco for ", false))
					}

				}
			case *slackevents.ReactionAddedEvent:
				{
					fmt.Println("Reaction", ev.Reaction)
					if strings.Contains(ev.Reaction, "taco") {
						api.PostMessage(ev.Item.Channel, slack.MsgOptionText("Yes, Here is a taco for ", false))
					}
				}
			}
		}
	})
	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":"+port, nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")
}

func main() {
	port := os.Getenv("PORT")
	slackBot(port)

}
