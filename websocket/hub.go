package websocket

import (
	"encoding/json"
	"log"
)

type Hub struct {
	// Mapa de clientes conectados
	clients map[*Client]bool
	// Canal para broadcast de mensagens
	BroadcastChan chan []byte
	// Canal para registrar novos clientes
	Register chan *Client
	// Canal para remover clientes
	Unregister chan *Client
	// Mapa de userID para clientes (para envio direcionado)
	userClients map[int][]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		BroadcastChan: make(chan []byte),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		userClients:   make(map[int][]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = true
			h.userClients[client.UserID] = append(h.userClients[client.UserID], client)
			log.Printf("Cliente registrado. UserID: %d", client.UserID)

		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.removeUserClient(client)
				log.Printf("Cliente desregistrado. UserID: %d", client.UserID)
			}

		case message := <-h.BroadcastChan:
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
					h.removeUserClient(client)
				}
			}
		}
	}
}

// SendToUser envia uma notificação para um usuário específico
func (h *Hub) SendToUser(userID int, notification interface{}) {
	data, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Erro ao serializar notificação: %v", err)
		return
	}

	log.Printf("SendToUser %d: %s", userID, string(data))

	if clients, ok := h.userClients[userID]; ok {
		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
				h.removeUserClient(client)
			}
		}
	}
}

func (h *Hub) removeUserClient(client *Client) {
	if clients, ok := h.userClients[client.UserID]; ok {
		newClients := make([]*Client, 0)
		for _, c := range clients {
			if c != client {
				newClients = append(newClients, c)
			}
		}
		if len(newClients) == 0 {
			delete(h.userClients, client.UserID)
		} else {
			h.userClients[client.UserID] = newClients
		}
	}
}

// Broadcast envia uma mensagem para todos os clientes conectados
func (h *Hub) Broadcast(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Erro ao serializar mensagem de broadcast: %v", err)
		return
	}

	log.Printf("Broadcast: %s", string(data))

	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clients, client)
			h.removeUserClient(client)
		}
	}
}
