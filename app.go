package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"net/http"
	"net/url"
)

type requestInfo struct {
	Path     string `json:"path"`
	Request  string `json:"request"`
	Response string `json:"response"`
	Ok       bool   `json:"ok"`
}

func main() {
	log.Printf("Starting server\n")

	http.Handle("/api/auth.test", route(authTest))
	http.Handle("/api/users.lookupByEmail", route(usersLookupByEmail))
	http.Handle("/api/im.open", route(imOpen))
	http.Handle("/api/chat.postMessage", route(chatPostMessage))

	log.Fatal(http.ListenAndServe(":9393", nil))
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
			res = `{"ok":false,"error":"not_authed"}`
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
	routeStub := regexp.MustCompile(`/api/([^\/]+)`).ReplaceAllString(info.Path, "$1")
	// e.g. /api/im.open -> im.open

	logFile, err := os.Create(fmt.Sprintf("/messages/slack/%s_%s", routeStub, ts))
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

// usersLookupByEmail succeeds always (remember, it won't be called if the actual auth check earlier doesn't succeed)
// required fields: <none>
func authTest(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	res = `{"ok":true}`
	ok = true
	return
}

// usersLookupByEmail succeeds iff values.email is a valid email
// required fields: email
func usersLookupByEmail(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	email := values.Get("email")
	if isEmail(email) {
		res = `{"ok":true,"user":{"id":"UXXXXXXXX"}}`
		ok = true
		return
	} else {
		res = `{"ok":false,"error":"users_not_found"}`
		ok = false
		return
	}
}

// imOpen succeeds iff values.user matches /U.{8}/
// required fields: user
func imOpen(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	user := values.Get("user")
	ok = user != "" && isSlackUser(user)
	if ok {
		res = `{"ok":true,"channel":{"id":"DXXXXXXXX"}}`
		return
	} else {
		res = `{"ok":false,"error":"user_not_found"}`
		return
	}
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

// chatPostMessage succeeds iff it has its required fields
// required fields: channel, text
// optional fields: username, attachments
func chatPostMessage(ts string, values url.Values, w http.ResponseWriter, r *http.Request) (res string, ok bool, err error) {
	channel := values.Get("channel")
	text := values.Get("text")
	username := stringOrDefault(values.Get("username"), "default-username")

	if channel == "" {
		res = `{"ok":false,"error":"channel_not_found"}`
		ok = false
		return
	}
	if text == "" {
		res = `{"ok":false,"error":"no_text"}`
		ok = false
		return
	}

	resStruct := chatPostMessageResponse{
		Ok:      true,
		Channel: channel,
		Ts:      ts,
		Message: chatPostMessageResponseMessage{
			Type:        "message",
			Subtype:     "bot_message",
			Text:        text,
			Ts:          ts,
			Username:    username,
			BotID:       "BXXXXXXXX",
			Attachments: values.Get("attachments"),
		},
	}
	resBytes, err := json.Marshal(resStruct)
	if err != nil {
		return
	}

	res = string(resBytes)
	ok = true
	return
}
