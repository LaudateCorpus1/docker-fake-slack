package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"net/http"
	"net/url"

	"github.com/nlopes/slack"
)

var (
	notAuthedResponse = `{ "ok": false, "error": "not_authed" }`
)

type slackResponse struct {
	Ok      bool                 `json:"ok"`
	Channel string               `json:"channel"`
	Ts      string               `json:"ts"`
	Message slackResponseMessage `json:"message"`
}
type slackResponseMessage struct {
	Type        string             `json:"type"`
	Subtype     string             `json:"subtype"`
	Text        string             `json:"text"`
	Ts          string             `json:"ts"`
	Username    string             `json:"username"`
	BotID       string             `json:"bot_id"`
	Attachments []slack.Attachment `json:"attachments,omitempty"`
}

type requestInfo struct {
	Path     string `json:"path"`
	Request  string `json:"request"`
	Response string `json:"response"`
	Ok       bool   `json:"ok"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	runServer()
	// testJSON()
}

func runServer() {
	log.Printf("Starting server\n")

	http.Handle("/api/auth.test", route(authTest))
	http.Handle("/api/users.lookupByEmail", route(usersLookupByEmail))
	http.Handle("/api/im.open", route(imOpen))
	http.Handle("/api/chat.postMessage", route(chatPostMessage))

	log.Fatal(http.ListenAndServe(":9393", nil))
}

// getTimestamp returns timestamps in the format the slack API uses
func getTimestamp() string {
	nanos := time.Now().UnixNano()
	ts := fmt.Sprintf("%.6f", float64(nanos)/math.Pow(10, 9))
	return ts
}

// a routerCore is similar to an http.HandlerFunc; wrap it with `route` to get an http.HandlerFunc
// params in:
//   ts: timestamp
//   req: the request body
// params out:
//   res: the response
//   ok: return true, normally. If the request is malformed (client error), return false
//   err: return nil, normally. If the server fails (server error), return the error
type routerCore func(ts string, req string, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error)

// route wraps a simple handler function in logging and auth-checking
func route(core routerCore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts := getTimestamp()
		log.Printf("[%s] %s\n", ts, r.URL.Path)

		reqBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			msg := fmt.Sprintf("Could not read body: %s", err.Error())
			log.Printf(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		req := string(reqBytes)

		var res string
		token := r.Header.Get("Authorization")
		ok := token != "" // extremely lenient authorization
		if ok {
			res, ok, err = core(ts, req, w, r)
			if err != nil {
				log.Printf(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			res = notAuthedResponse
		}
		fmt.Fprintf(w, res)

		logRequest(w, ts, requestInfo{
			Path:     r.URL.Path,
			Request:  req,
			Response: res,
			Ok:       ok,
		})
	})
}

// logs a request + response to disk at /messages/slack/request_{{ts}}
func logRequest(w http.ResponseWriter, ts string, info requestInfo) {
	logFile, err := os.Create(fmt.Sprintf("/messages/slack/%s", ts))
	if err != nil {
		msg := fmt.Sprintf("Could not create logfile: %s", err.Error())
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer logFile.Close()

	bytes, err := json.Marshal(info)
	if err != nil {
		msg := fmt.Sprintf("Could not build JSON log line: %s", err.Error())
		log.Printf(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	logFile.Write(bytes)
	logFile.WriteString("\n")
}

//
// route handlers:
//

func authTest(ts string, req string, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	ok = true
	res = `{ "ok": true }`
	return
}

func usersLookupByEmail(ts string, req string, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	res = `{ "ok": true, "user": { "id": "UXXXXXXXX" } }`
	ok = true
	return
}

func imOpen(ts string, req string, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	res = `{ "ok": true, "channel": { "id": "DXXXXXXXX" } }`
	ok = true
	return
}

func chatPostMessage(ts string, req string, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	// @TODO check for content-type application/json header?

	values, err := url.ParseQuery(req)
	if err != nil {
		ok = false
		return
	}
	// fmt.Printf("%v\n", values)
	_ = values

	//
	// respond to the request
	//

	resStruct := slackResponse{
		Ok:      true,
		Channel: "TODO",
		Ts:      ts,
		Message: slackResponseMessage{
			Type:     "message",
			Subtype:  "bot_message",
			Text:     "TODO",
			Ts:       ts,
			Username: "TODO",
			BotID:    "TODO",
			// Attachments []slack.Attachment // @TODO
		},
	}
	resBytes, err := json.Marshal(resStruct)
	if err != nil {
		ok = false
		return
	}
	res = string(resBytes)
	w.Header().Set("Content-Type", "application/json")
	return
}
