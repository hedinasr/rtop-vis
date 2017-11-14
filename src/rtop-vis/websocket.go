package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type msg struct {
	Num int
}

func startWS() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", rootHandler)

	panic(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ = template.ParseFiles("index.html")

	if err := tmpl.Execute(w, allStats); err != nil {
		log.Print(err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Client connected")

	//if r.Header.Get("Origin") != "http://"+r.Host {
	//	http.Error(w, "Origin not allowed", 403)
	//	return
	//}

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	quit := make(chan int)
	go writeStats(conn, quit)
	go checkClosed(conn, quit)
}

// Send stats to the client.
func writeStats(conn *websocket.Conn, quit chan int) {
	for {
		select {
		// If checkClosed function send us the quit chan, then return
		case <-quit:
			return
		case stat := <-allStats:
			if err := conn.WriteJSON(stat); err != nil {
				fmt.Println("Write error", err)
			}
		}
	}
}

// Check if client has closed the connection.
func checkClosed(conn *websocket.Conn, quit chan int) {
	defer func() {
		quit <- 1
	}()
	m := msg{}

	err := conn.ReadJSON(&m)
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
			fmt.Println("Websocket closed!")
		}
		fmt.Println("Error reading json.", err)
	}

	fmt.Printf("Got message: %#v\n", m)
}
