package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	que "github.com/bgentry/que-go"
	qe "github.com/heroku-examples/go-queue-example"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

var (
	log     = logrus.WithField("cmd", "queue-example-web")
	qc      *que.Client
	pgxpool *pgx.ConnPool
)

// queueIndexRequest into the que as an encoded JSON object
func queueIndexRequest(ir qe.IndexRequest) error {
	enc, err := json.Marshal(ir)
	if err != nil {
		return errors.Wrap(err, "Marshalling the IndexRequest")
	}

	j := que.Job{
		Type: qe.IndexRequestJob,
		Args: enc,
	}

	return errors.Wrap(qc.Enqueue(&j), "Enqueueing Job")
}

// getIndexRequest from the body and further validate it.
func getIndexRequest(r io.Reader) (qe.IndexRequest, error) {
	var ir qe.IndexRequest
	rd := json.NewDecoder(r)
	if err := rd.Decode(&ir); err != nil {
		return ir, errors.Wrap(err, "Error decoding JSON body.")
	}

	if ir.URL == "" || !strings.HasPrefix(ir.URL, "http") {
		return ir, errors.New("The request did not contain a url or was invalid")
	}

	_, err := url.Parse(ir.URL)
	if err != nil {
		return ir, errors.Wrap(err, "Error parsing URL")
	}

	return ir, nil
}

// handlePostIndexRequest from the outside world. We validate the request and
// enqueue it for later processing returning a 202 if there were no errors
func handlePostIndexRequest(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("func", "handlePostIndexRequest")
	ir, err := getIndexRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		l.Println(err.Error())
		return
	}

	if err := queueIndexRequest(ir); err != nil {
		l.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func handleIndexRequest(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("func", "handleIndexRequest")
	switch r.Method {
	case "POST":
		handlePostIndexRequest(w, r)
	default:
		err := "Invalid http method. Only POST is accepted."
		l.WithField("method", r.Method).Println(err)
		http.Error(w, err, http.StatusBadRequest)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("func", "handleIndex")
	if _, err := io.WriteString(w, `Usage: curl -XPOST "https://<app name>.herokuapp.com/index" -d '{"url": "http://google.com"}'`); err != nil {
		l.Println(err.Error())
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	var err error
	pgxpool, qc, err = qe.Setup(dbURL)
	if err != nil {
		log.WithField("DATABASE_URL", dbURL).Fatal("Unable to setup queue / database")
	}
	defer pgxpool.Close()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/index", handleIndexRequest)
	log.Println(http.ListenAndServe(":"+port, nil))
}
