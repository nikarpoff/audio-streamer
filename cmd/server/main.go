package main

import (
	"flag"
	"log"
	"net"

	"github.com/nikarpoff/audio-streamer/internal/network"
)

var addr = flag.String("addr", ":7001", "udp service address")

func main() {
	flag.Parse()

	udpAddr, err := net.ResolveUDPAddr("udp", *addr)
	if err != nil {
		log.Fatal("Error occuired while resolving UDP address:", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("ListenUDP:", err)
	}
	defer conn.Close()

	hub := network.NewHub()
	go hub.Run(conn)

	log.Println("UDP server starting on", *addr)

	buffer := make([]byte, 2048)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("ReadFromUDP error:", err)
			continue
		}
		hub.Register(clientAddr)

		packet := make([]byte, n)
		copy(packet, buffer[:n])
		hub.Broadcast(network.AudioMessage{Sender: clientAddr, Data: packet})
	}
}
