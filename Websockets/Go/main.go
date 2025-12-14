package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// ChatServer manages WebSocket connections
type ChatServer struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (server *ChatServer) run() {
	for {
		select {
		case client := <-server.register:
			server.clients[client] = true
			fmt.Printf("Client connected. Total: %d\n", len(server.clients))

		case client := <-server.unregister:
			if _, ok := server.clients[client]; ok {
				delete(server.clients, client)
				client.Close()
				fmt.Printf("Client disconnected. Total: %d\n", len(server.clients))
			}

		case message := <-server.broadcast:
			for client := range server.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Error writing message: %v", err)
					server.unregister <- client
				}
			}
		}
	}
}

func (server *ChatServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	defer func() {
		server.unregister <- conn
	}()

	server.register <- conn

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if len(message) > 0 {
			log.Printf("Received message: %s", string(message))
			server.broadcast <- message
		}
	}
}

func main() {
	server := NewChatServer()
	go server.run()

	http.HandleFunc("/ws", server.handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir("./static/")))

	log.Println("Chat server starting on 0.0.0.0:8090...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8090", nil))
}
