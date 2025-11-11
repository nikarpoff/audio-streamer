package network

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan AudioMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// Audio Message includes message data and who send this
type AudioMessage struct {
	sender *Client
	data   []byte
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan AudioMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		// We got new client to register?
		case client := <-h.register:
			h.clients[client] = true
		// We got old client to unregister?
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		// We got new broadcast message?
		case message := <-h.broadcast:
			for client := range h.clients {
				if client == message.sender {
					continue
				}

				select {
				case client.send <- message.data:
				default:
					// Client is unavailable
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
