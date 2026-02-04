// file path: internal/server/websocket.go
package server

import (
	"time"

	"CHIMERA/internal/models"
)

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

// Run: Hub'in ana döngüsü — register/unregister, broadcast ve hedefli gönderimleri işler
func (h *Hub) Run() {
	for {
		select {
		case <-h.Stop:
			log.Info("stopping", "component", "HUB")
			for _, c := range h.Clients {
				close(c.Send)
			}
			h.Clients = map[string]*Client{}
			return

		case c := <-h.Register:
			h.Clients[c.ID] = c
			log.Info("register", "component", "HUB", "client_id", c.ID, "clients_count", len(h.Clients))

		case c := <-h.Unregister:
			if _, ok := h.Clients[c.ID]; ok {
				close(c.Send)
				delete(h.Clients, c.ID)
				log.Info("unregister", "component", "HUB", "client_id", c.ID, "clients_count", len(h.Clients))
			}

		case env := <-h.Broadcast:
			for _, cl := range h.Clients {
				select {
				case cl.Send <- env:
				default:
					// kanal doluysa mesaj atla (drop) uyarrrıı
					log.Info("broadcast drop", "component", "HUB", "client_id", cl.ID, "reason", "channel full")
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
