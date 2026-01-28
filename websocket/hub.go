package websocket

import (
	"encoding/json"
	"log"
	"slices"
	"time"

	"github.com/xlzd/gotp"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	Type     string `json:"type"` // "handshake_attempt", "dm", "broadcast"
	SenderId string `json:"sender_id"`
	TargetId string `json:"target_id,omitempty"` // Only needed for DM or handshake
	Content  string `json:"content,omitempty"`   // The text message
	Code     string `json:"code,omitempty"`      // The 6-digit TOTP code
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) findClientByID(targetID string) *Client {
	for client := range h.clients {
		if client.id == targetID {
			return client
		}
	}
	return nil // Client not found
}

// client struct requries allow list to be added!
func (h *Hub) handleMessage(raw []byte) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	switch msg.Type {

	case "handshake_attempt":
		target := *h.findClientByID(msg.TargetId)

		isValid := gotp.NewDefaultTOTP(target.secret).Verify(msg.Code, time.Now().Unix())

		if isValid {
			target.allow = append(target.allow, msg.SenderId)
			sender := *h.findClientByID(msg.SenderId)
			sender.allow = append(sender.allow, msg.TargetId)
		} else {
			log.Printf("Direct messaging request denied.")
		}

	case "dm":
		targetClient := h.findClientByID(msg.TargetId)
		if targetClient == nil {
			return
		}
		isAuthorized := slices.Contains(targetClient.allow, msg.SenderId)

		if isAuthorized {
			// send message
		}
	default:
		for client := range h.clients {
			select {
			case client.send <- raw:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}
