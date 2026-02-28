package network

import (
	"net"
	"time"
)

const (
	clientSendBufferSize = 8
	clientIdleTimeout    = 30 * time.Second
	cleanupPeriod        = 5 * time.Second
	broadcastBufferSize  = 256
)

type udpClient struct {
	addr     *net.UDPAddr
	send     chan []byte
	lastSeen time.Time
}

// Audio Message includes message data and who send this
type audioMessage struct {
	sender *net.UDPAddr
	data   []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
type Hub struct {
	clients map[string]*udpClient

	broadcast  chan audioMessage
	register   chan *net.UDPAddr
	unregister chan string
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*udpClient),
		broadcast:  make(chan audioMessage, broadcastBufferSize),
		register:   make(chan *net.UDPAddr, 8),
		unregister: make(chan string, 8),
	}
}

func (h *Hub) Register(client *net.UDPAddr) {
	h.register <- client
}

func (h *Hub) Broadcast(sender *net.UDPAddr, packet *[]byte) {
	message := audioMessage{sender: sender, data: *packet}

	select {
	case h.broadcast <- message:
		// alright
	default:
		// Hub is overloaded. Drop the oldest queued packet to avoid blocking UDP reads
		// and prioritize fresh audio frames.
		select {
		case <-h.broadcast:
		default:
		}

		select {
		case h.broadcast <- message:
		default:
			// Queue is still full, skip this frame.
		}
	}
}

func (h *Hub) Run(conn *net.UDPConn) {
	cleanupTicker := time.NewTicker(cleanupPeriod)
	defer cleanupTicker.Stop()

	for {
		select {
		// We got new client to register?
		case client := <-h.register:
			h.registerClient(client, conn)
		// We got unregister client request?
		case key := <-h.unregister:
			h.removeClient(key)
		// We got new broadcast message?
		case message := <-h.broadcast:
			h.broadcastPacket(message)
		// Cleanup inactive clients if required!
		case <-cleanupTicker.C:
			h.cleanupInactiveClients()
		}
	}
}

func (h *Hub) registerClient(addr *net.UDPAddr, conn *net.UDPConn) {
	key := addr.String()
	if existing, ok := h.clients[key]; ok {
		existing.lastSeen = time.Now()
		return
	}

	client := &udpClient{
		addr:     addr,
		send:     make(chan []byte, clientSendBufferSize),
		lastSeen: time.Now(),
	}
	h.clients[key] = client

	go h.writePump(conn, key, client)
}

func (h *Hub) broadcastPacket(message audioMessage) {
	senderKey := ""
	if message.sender != nil {
		senderKey = message.sender.String()
	}

	for key, client := range h.clients {
		if key == senderKey {
			continue
		}

		select {
		case client.send <- message.data:
		default:
			// Slow client: drop packet to keep low latency and prevent global stalls.
		}
	}
}

func (h *Hub) cleanupInactiveClients() {
	now := time.Now()
	for key, client := range h.clients {
		if now.Sub(client.lastSeen) > clientIdleTimeout {
			h.removeClient(key)
		}
	}
}

func (h *Hub) removeClient(key string) {
	client, ok := h.clients[key]
	if !ok {
		return
	}

	delete(h.clients, key)
	close(client.send)
}

func (h *Hub) writePump(conn *net.UDPConn, key string, client *udpClient) {
	for packet := range client.send {
		if _, err := conn.WriteToUDP(packet, client.addr); err != nil {
			h.unregister <- key
			return
		}
	}
}
