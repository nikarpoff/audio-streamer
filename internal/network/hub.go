package network

import (
	"net"
)

// Audio Message includes message data and who send this
type AudioMessage struct {
	Sender *net.UDPAddr
	Data   []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
type Hub struct {
	clients map[string]*net.UDPAddr

	broadcast chan AudioMessage
	register  chan *net.UDPAddr
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[string]*net.UDPAddr),
		broadcast: make(chan AudioMessage, 1024),
		register:  make(chan *net.UDPAddr, 1024),
	}
}

func (h *Hub) Register(client *net.UDPAddr) {
	h.register <- client
}

func (h *Hub) Broadcast(message AudioMessage) {
	h.broadcast <- message
}

func (h *Hub) Run(conn *net.UDPConn) {
	for {
		select {
		// We got new client to register?
		case client := <-h.register:
			h.clients[client.String()] = client
		// We got new broadcast message?
		case message := <-h.broadcast:
			for key, client := range h.clients {
				// Don't send message to sender
				// if message.Sender != nil && key == message.Sender.String() {
				// 	continue
				// }

				// Send message to client
				if _, err := conn.WriteToUDP(message.Data, client); err != nil {
					delete(h.clients, key)
				}
			}
		}
	}
}
