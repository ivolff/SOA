package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

type Player struct {
	ID       string `json:"ID"`
	Username string `json:"Username"`
	Avatar   string `json:"Avatar"`
	Sex      string `json:"Sex"`
	Email    string `json:"Email"`
}

type GameStats struct {
	IsWin    bool `json:"IsWin,string"`
	Duration int  `json:"Duration,string"`
}

type QueueMsg struct {
	ID        string `json:"ID"`
	Username  string `json:"Username"`
	Avatar    string `json:"Avatar"`
	Sex       string `json:"Sex"`
	Email     string `json:"Email"`
	WinCount  int    `json:"WinCount"`
	LossCount int    `json:"LossCount"`
	Duration  int    `json:"Duration"`
	Filename  string `json:"Filename"`
}

var (
	db    *sql.DB
	ch    *amqp.Channel
	queue amqp.Queue
)

const path_to_store = "store"

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func notifyOnError(err error, msg string) bool {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		return true
	}
	return false
}

func addPlayer(w http.ResponseWriter, r *http.Request) {
	var newPlayer Player
	reqBody, err := ioutil.ReadAll(r.Body)
	notifyOnError(err, "")

	json.Unmarshal(reqBody, &newPlayer)
	newPlayer.ID = uuid.New().String()

	err = dbAddPlayer(db, &newPlayer)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newPlayer)
}

func updatePlayer(w http.ResponseWriter, r *http.Request) {
	var newPlayer Player
	reqBody, err := ioutil.ReadAll(r.Body)
	notifyOnError(err, "")

	json.Unmarshal(reqBody, &newPlayer)

	newPlayer.ID = mux.Vars(r)["id"]

	dbUpdatePlayer(db, &newPlayer)

	w.WriteHeader(http.StatusAccepted)
}

func getAllPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := dbGetAllPlayers(db)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(players)
}

func getOnePlayer(w http.ResponseWriter, r *http.Request) {
	playerID := mux.Vars(r)["id"]

	player, err := dbGetOnePlayer(db, playerID)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(player)
}

func updatePlayerStats(w http.ResponseWriter, r *http.Request) {
	playerID := mux.Vars(r)["id"]

	var stats GameStats
	reqBody, err := ioutil.ReadAll(r.Body)
	notifyOnError(err, "")

	err = json.Unmarshal(reqBody, &stats)
	failOnError(err, "")

	err = dbUpdatePlayerStats(db, &stats, playerID)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func getPlayerStats(w http.ResponseWriter, r *http.Request) {
	log.Print(mux.Vars(r))
	playerID := mux.Vars(r)["id"]

	msg, err := dbGetPlayerStats(db, playerID)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(msg)
	failOnError(err, "")

	err = ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		},
	)
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(r.Host + "/players/" + playerID + "/stats/" + msg.Filename))
}

func getPdf(w http.ResponseWriter, r *http.Request) {
	playerID := mux.Vars(r)["id"]
	filename := mux.Vars(r)["filename"]

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	fw, err := writer.CreateFormFile("pdf", filename)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, err := os.Open(path_to_store + "/pdfs/" + playerID + filename)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = io.Copy(fw, file)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", writer.FormDataContentType())
	w.Write(body.Bytes())
	w.WriteHeader(http.StatusOK)
}

func storePdf(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if notifyOnError(err, "") {
		return
	}

	filePath := path_to_store + "/pdfs/" + mux.Vars(r)["id"] + mux.Vars(r)["filename"]

	err = os.WriteFile(filePath, body, 0666)
	if notifyOnError(err, "") {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	db = connectToBD()
	defer db.Close()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	queue, err = ch.QueueDeclare(
		"work_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/players", addPlayer).Methods("POST")
	router.HandleFunc("/players", getAllPlayers).Methods("GET")
	router.HandleFunc("/players/{id}", updatePlayer).Methods("PUT")
	router.HandleFunc("/players/{id}", getOnePlayer).Methods("GET")
	router.HandleFunc("/players/{id}/stats", updatePlayerStats).Methods("PUT")
	router.HandleFunc("/players/{id}/stats", getPlayerStats).Methods("GET")
	router.HandleFunc("/players/{id}/stats/{filename}", getPdf).Methods("GET")
	router.HandleFunc("/players/{id}/stats/{filename}", storePdf).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
