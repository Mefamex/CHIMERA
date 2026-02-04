// file path: internal/server/structs.go
package server

import (
	"CHIMERA/internal/core"
	"CHIMERA/internal/models"
)

// server globals
var log = core.Log

// global hub
var HubInstance = NewHub()

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
