package main

import (
	"net"
	"encoding/hex"
	"log"
)

const socketBufferSize = 1460

// client represents the local machine sending the DVR stream
type client struct {
	// conn is the TCP connection
	conn *net.TCPConn
	// send is the channel on which messages are sent
	send chan []byte
	// channel is the channel number of the stream
	channel *int
}

// newClient makes a new client
func newClient(channel *int) *client {
	return &client{
		send:    make(chan []byte, socketBufferSize),
		channel: channel,
	}
}

// handle handles events such as messages being sent
func handle(c *client) {
	for {
		select {
		case message := <-c.send:
			// TODO: Send this to the server
			log.Printf("Sent:\n%v", hex.Dump(message))
		}
	}
}
