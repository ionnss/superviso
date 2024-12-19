package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait                    = 10 * time.Second
	pongWait                     = 60 * time.Second
	pingPeriod                   = (pongWait * 9) / 10
	maxMessageSize               = 512
	MessageTypeSlotUpdate        = "slot_update"
	MessageTypeAppointmentUpdate = "appointment_update"
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID int
}

type SlotUpdateMessage struct {
	Type         string `json:"type"`
	SlotID       int    `json:"slot_id"`
	Status       string `json:"status"`
	SupervisorID int    `json:"supervisor_id"`
	Date         string `json:"date"`
	StartTime    string `json:"start_time"`
}

type AppointmentUpdateMessage struct {
	Type          string `json:"type"`
	AppointmentID int    `json:"appointment_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erro na leitura do websocket: %v", err)
			} else {
				log.Printf("WebSocket fechado normalmente. UserID: %d", c.UserID)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
