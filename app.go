package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"
)

type RequestInfo struct {
	Path string `json:"path"`
	Body string `json:"body"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	log.Printf("Starting server\n")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "Method, %s\n", r.Method)
		// fmt.Fprintf(w, "Path %s\n", r.URL.Path)

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()

		nanos := time.Now().UnixNano()
		fmt.Fprintf(w, "%d: %s\n", nanos, body)

		f, err := os.Create(fmt.Sprintf("/messages/slack/request_%d", nanos))
		check(err)
		defer f.Close()

		reqInfo := RequestInfo{
			Path: r.URL.Path,
			Body: body,
		}
		bytes, err := json.Marshal(reqInfo)
		check(err)

		f.Write(bytes)
		f.WriteString("\n")
	})

	log.Fatal(http.ListenAndServe(":9393", nil))
}
