package network

import (
	"net"
	"testing"
	"time"
)

func TestHubBroadcastSkipsSenderAndDeliversToOtherClient(t *testing.T) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("resolve server addr: %v", err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("listen server: %v", err)
	}
	defer serverConn.Close()

	hub := NewHub()
	go hub.Run(serverConn)

	listenerA, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen client A: %v", err)
	}
	defer listenerA.Close()

	listenerB, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("listen client B: %v", err)
	}
	defer listenerB.Close()

	hub.Register(listenerA.LocalAddr().(*net.UDPAddr))
	hub.Register(listenerB.LocalAddr().(*net.UDPAddr))

	payload := []byte{1, 2, 3, 4}
	hub.Broadcast(listenerA.LocalAddr().(*net.UDPAddr), &payload)

	_ = listenerB.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 16)
	n, _, err := listenerB.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read from client B: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("unexpected payload size: got %d want %d", n, len(payload))
	}

	_ = listenerA.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if _, _, err := listenerA.ReadFromUDP(buf); err == nil {
		t.Fatalf("sender received its own packet")
	}
}
