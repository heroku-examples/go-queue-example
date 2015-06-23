package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	qe "github.com/heroku-examples/go_queue_example"

	log "github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/bgentry/que-go"
	"github.com/heroku-examples/go_queue_example/Godeps/_workspace/src/github.com/codegangsta/negroni"
)

// queueIndexRequest into the que as an encoded JSON object
func queueIndexRequest(ir qe.IndexRequest) error {
	enc, err := json.Marshal(ir)
	if err != nil {
		return err
	}

	j := que.Job{
		Type: qe.IndexRequestJob,
		Args: enc,
	}

	return qe.Qc.Enqueue(&j)
}

// getIndexRequest from the body and further validate it.
func getIndexRequest(r io.Reader) (qe.IndexRequest, error) {
	var ir qe.IndexRequest
	rd := json.NewDecoder(r)
	if err := rd.Decode(&ir); err != nil {
		return ir, fmt.Errorf("Error decoding JSON body.")
	}

	if ir.URL == "" || !strings.HasPrefix(ir.URL, "http") {
		return ir, fmt.Errorf("The request did not contain a url or was invalid")
	}

	_, err := url.Parse(ir.URL)
	if err != nil {
		return ir, fmt.Errorf("Error parsing URL: %s", err.Error())
	}

	return ir, nil
}

// handlePostIndexRequest from the outside world. We validate the request and
// enqueue it for later processing returning a 202 if there were no errors
func handlePostIndexRequest(w http.ResponseWriter, r *http.Request) {
	ir, err := getIndexRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := queueIndexRequest(ir); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func handleGetIndexStatus(w http.ResponseWriter, r *http.Request) {
}

func handleIndexRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetIndexStatus(w, r)
	case "POST":
		handlePostIndexRequest(w, r)
	default:
		http.Error(w, "Invalud http method. Only GET & POST accepted.", http.StatusBadRequest)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	defer qe.PgxPool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/index", handleIndexRequest)

	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + port)
}
