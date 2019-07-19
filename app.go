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
)

var (
	notAuthedResponse = `{ "ok": false, "error": "not_authed" }`
)

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
//   values: the values, parsed from the request body
// params out:
//   res: the response
//   ok: return true, normally. If the request is malformed (client error), return false
//   err: return nil, normally. If the server fails (server error), return the error
type routerCore func(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error)

// route wraps a simple handler function in logging and auth-checking
func route(core routerCore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts := getTimestamp()
		log.Printf("[%s] %s %s\n", ts, r.Method, r.URL.Path)

		//
		// parse body
		//

		if r.Header.Get("Content-type") != "application/x-www-form-urlencoded" {
			msg := fmt.Sprintf("only 'Content-type: application/x-www-form-urlencoded' is accepted (%s is no good)", r.Header.Get("Content-type"))
			log.Printf(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		reqBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			msg := fmt.Sprintf("Could not read body: %s", err.Error())
			log.Printf(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		req := string(reqBytes)

		values, err := url.ParseQuery(req)
		if err != nil {
			msg := fmt.Sprintf("Could not parse body values: %s", err.Error())
			log.Printf(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		//
		// get response from core
		//

		var res string
		ok := values.Get("token") != "" // extremely lenient authorization
		if ok {
			res, ok, err = core(ts, values, w, r)
			if err != nil {
				log.Printf(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			res = notAuthedResponse
		}

		//
		// send response
		//

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, res)

		//
		// log request
		//

		// note that these loggers will also log the token
		// but that's fine since you shouldn't be sending real tokens to this dummy service anyway
		log.Printf("[%s] %s %s %s\n", ts, r.Method, r.URL.Path, values)
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

func authTest(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	res = `{ "ok": true }`
	ok = true
	return
}

func usersLookupByEmail(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	res = `{ "ok": true, "user": { "id": "UXXXXXXXX" } }`
	ok = true
	return
}

func imOpen(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	res = `{ "ok": true, "channel": { "id": "DXXXXXXXX" } }`
	ok = true
	return
}

type chatPostMessageResponse struct {
	Ok      bool                           `json:"ok"`
	Channel string                         `json:"channel"`
	Ts      string                         `json:"ts"`
	Message chatPostMessageResponseMessage `json:"message"`
}
type chatPostMessageResponseMessage struct {
	Type        string `json:"type"`
	Subtype     string `json:"subtype"`
	Text        string `json:"text"`
	Ts          string `json:"ts"`
	Username    string `json:"username"`
	BotID       string `json:"bot_id"`
	Attachments string `json:"attachments,omitempty"`
}

func stringDefault(s, fallback string) string {
	if s == "" {
		return fallback
	} else {
		return s
	}
}

func chatPostMessage(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	resStruct := chatPostMessageResponse{
		Ok:      true,
		Channel: values.Get("channel"),
		Ts:      ts,
		Message: chatPostMessageResponseMessage{
			Type:        "message",
			Subtype:     "bot_message",
			Text:        values.Get("text"),
			Ts:          ts,
			Username:    stringDefault(values.Get("username"), "default-username"),
			BotID:       "BXXXXXXXX",
			Attachments: values.Get("attachments"),
		},
	}
	resBytes, err := json.Marshal(resStruct)
	if err != nil {
		ok = false
		return
	}
	res = string(resBytes)
	return
}
