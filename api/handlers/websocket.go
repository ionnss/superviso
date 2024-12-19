package handlers

import (
	"log"
	"net/http"
	"superviso/api/auth"
	"superviso/websocket"

	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Em produção, configurar origem permitida
	},
}

func WebSocketHandler(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		log.Printf("Nova conexão WebSocket. UserID: %d, RemoteAddr: %s", userID, r.RemoteAddr)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Erro ao fazer upgrade da conexão: %v", err)
			return
		}

		client := &websocket.Client{
			Hub:    hub,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			UserID: userID,
		}
		hub.Register <- client

		// Iniciar goroutines de leitura/escrita
		go client.WritePump()
		go client.ReadPump()
	}
}
