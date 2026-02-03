// file path: internal/server/websocket.go
package server

import (
	"log"
	"time"

	"CHIMERA/internal/models"
)

// Client temsilcisi: tek bir websocket bağlantısının gönderim kanalı
type Client struct {
	ID   string
	Send chan models.Envelope
}

// directEnvelope: tek bir client'a gönderim için paket
type directEnvelope struct {
	ID  string
	Env models.Envelope
}

// Hub: client kaydı, broadcast ve direkt gönderimler için merkez
type Hub struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[string]*Client
	Broadcast  chan models.Envelope
	Send       chan directEnvelope
	Stop       chan struct{}
}

// global hub örneği
var HubInstance = NewHub()

func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan models.Envelope),
		Send:       make(chan directEnvelope),
		Stop:       make(chan struct{}),
	}
}

// Run: Hub'in ana döngüsü — register/unregister, broadcast ve hedefli gönderimleri işler
func (h *Hub) Run() {
	log.Println("hub: started")
	for {
		select {
		case <-h.Stop:
			log.Println("hub: stopping")
			for _, c := range h.Clients {
				close(c.Send)
			}
			h.Clients = map[string]*Client{}
			return

		case c := <-h.Register:
			h.Clients[c.ID] = c
			log.Printf("hub: register %s (clients=%d)\n", c.ID, len(h.Clients))

		case c := <-h.Unregister:
			if _, ok := h.Clients[c.ID]; ok {
				close(c.Send)
				delete(h.Clients, c.ID)
				log.Printf("hub: unregister %s (clients=%d)\n", c.ID, len(h.Clients))
			}

		case env := <-h.Broadcast:
			for _, cl := range h.Clients {
				select {
				case cl.Send <- env:
				default:
					// kanal doluysa mesaj atla (drop)
				}
			}

		case de := <-h.Send:
			if cl, ok := h.Clients[de.ID]; ok {
				select {
				case cl.Send <- de.Env:
					// sent
				case <-time.After(2 * time.Second):
					// still blocked, drop
				}
			}
		}
	}
}

// Yardımcılar
func (h *Hub) BroadcastEnvelope(env models.Envelope) {
	h.Broadcast <- env
}

func (h *Hub) SendToClient(id string, env models.Envelope) {
	h.Send <- directEnvelope{ID: id, Env: env}
}

func (h *Hub) Shutdown() {
	close(h.Stop)
}
