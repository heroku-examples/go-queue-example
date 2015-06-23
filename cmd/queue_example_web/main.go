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

	log "github.com/Sirupsen/logrus"
	"github.com/bgentry/que-go"
	"github.com/codegangsta/negroni"
)

var (
	qc *que.Client
)

type indexRequest struct {
	URL string `json:url`
}

// queueIndexRequest into the que as an encoded JSON object
func queueIndexRequest(ir indexRequest) error {
	enc, err := json.Marshal(ir)
	if err != nil {
		return err
	}

	j := que.Job{
		Type: qe.IndexRequestJob,
		Args: enc,
	}

	return qc.Enqueue(&j)
}

// getIndexRequest from the body and further validate it.
func getIndexRequest(r io.Reader) (indexRequest, error) {
	var ir indexRequest
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
	if port != "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	pgxpool, err := qe.GetPgxPool(dbURL)
	if err != nil {
		log.WithField("DATABASE_URL", dbURL).Fatal(err)
	}
	defer pgxpool.Close()

	qc = que.NewClient(pgxpool)

	mux := http.NewServeMux()
	mux.HandleFunc("/index", handleIndexRequest)

	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + port)
}
