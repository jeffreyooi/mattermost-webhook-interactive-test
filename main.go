package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	mattermostAddr = "http://127.0.0.1:8065/hooks/9h5ifb7ofp8fdno7ejgefai91c"
)

type message struct {
	Text        string        `json:"text,omitempty"`
	Channel     string        `json:"channel,omitempty"`
	Attachments []*attachment `json:"attachments,omitempty"`
}

type attachment struct {
	Text    string    `json:"text,omitempty"`
	Actions []*action `json:"actions,omitempty"`
}

type action struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name,omitempty"`
	Integration *integration `json:"integration,omitempty"`
}

type integration struct {
	// if id field is emoty, the request URL suffix will be hash instead of id
	// for example
	// with id: http://127.0.0.1:8065/api/v4/posts/ea983ajd5byhb8opxtqbffwatw/actions/ack
	// without id: http://127.0.0.1:8065/api/v4/posts/ea983ajd5byhb8opxtqbffwatw/actions/gjn3kwwk3i8p38escnwcnrta4w
	// from initial testing, it seems like it doesn't have any side effects or error if this is set or not,
	// so this field can be safely ignored
	ID string `json:"id,omitempty"`
	// URL of the service that handles request fired from Mattermost server
	URL     string   `json:"url,omitempty"`
	Context *context `json:"context,omitempty"`
}

// Context is the payload that is set on each action in the message.
// The same context corresponding to the action will be sent to the
// service handling the callback request
type context struct {
	Action   string `json:"action,omitempty"`
	Duration int64  `json:"duration,omitempty"`
}

/*
	Sample response sent by Mattermost server:
	{
    "user_id": "x43hpmisdbna5ehyktio81fczw",
    "user_name": "jeffrey_ooi",
    "channel_id": "8kbh48gmh7fq7y87tz43m4kkqr",
    "channel_name": "test",
    "team_id": "k6xgzrpeiigypnur3w6k91868y",
    "team_domain": "test",
    "post_id": "ea983ajd5byhb8opxtqbffwatw",
    "trigger_id": "some_very_long_trigger_id_that_i_replaced",
    "type": "",
    "data_source": "",
    "context": {
        "action": "ack"
    }
}
*/
// The request sent by Mattermost server when user clicks on
// an interactive button
type callback struct {
	UserID      string   `json:"user_id,omitempty"`
	UserName    string   `json:"user_name,omitempty"`
	ChannelID   string   `json:"channel_id,omitempty"`
	ChannelName string   `json:"channel_name,omitempty"`
	TeamID      string   `json:"team_id,omitempty"`
	TeamDomain  string   `json:"team_domain,omitempty"`
	PostID      string   `json:"post_id,omitempty"`
	TriggerID   string   `json:"trigger_id,omitempty"`
	Type        string   `json:"type,omitempty"`
	DataSource  string   `json:"data_source,omitempty"`
	Context     *context `json:"context,omitempty"`
}

type update struct {
	Update *msg `json:"update,omitempty"`
}

type msg struct {
	Message string `json:"message,omitempty"`
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var c callback
	if err := json.Unmarshal(body, &c); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u := &update{
		Update: &msg{},
	}

	switch c.Context.Action {
	case "ack":
		u.Update.Message = fmt.Sprintf("acknowledged, investigating by @%s", c.UserName)
	case "ignore":
		u.Update.Message = fmt.Sprintf("ignore for %d, verified by @%s", c.Context.Duration, c.UserName)
	}

	buf, _ := json.Marshal(&u)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")

	msg := &message{
		Text:    text,
		Channel: "test",
		Attachments: []*attachment{
			{
				Text: text,
				Actions: []*action{
					{
						Name: "Acknowledge",
						Integration: &integration{
							ID:  "ack",
							URL: "http://192.168.1.7:8080/api/callback",
							Context: &context{
								Action: "ack",
							},
						},
					},
					{
						Name: "Ignore",
						Integration: &integration{
							ID:  "ignore",
							URL: "http://192.168.1.7:8080/api/callback",
							Context: &context{
								Action:   "ignore",
								Duration: 60,
							},
						},
					},
				},
			},
		},
	}

	buf, _ := json.Marshal(msg)
	bytesBuf := bytes.NewBuffer(buf)

	if resp, err := http.Post(mattermostAddr, "application/json", bytesBuf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err.Error())
	} else {
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err.Error())
		} else if string(body) != "ok" {
			fmt.Println(string(body))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(body)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func main() {
	m := mux.NewRouter()

	m.HandleFunc("/api/send", handleSend).Methods(http.MethodPost)
	m.HandleFunc("/api/callback", handleCallback).Methods(http.MethodPost)

	if err := http.ListenAndServe("0.0.0.0:8080", m); err != nil {
		panic(err)
	}
}
