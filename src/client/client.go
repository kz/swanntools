package main

import (
	"net"
	"encoding/hex"
	"log"
	"fmt"
	"os"
	"crypto/x509"
	"crypto/tls"
	"io/ioutil"
)

const socketBufferSize = 1460

// client represents the local machine sending the DVR stream
type client struct {
	// conn is the TCP (w/ TLS) connection to the server
	conn *tls.Conn
	// send is the channel on which messages are sent
	send chan []byte
	// channel is the channel number of the stream
	channel *int
}

// Client makes a new client
func Client(channel int) *client {
	// TODO: Initialize a TCP connection with the server and send a message with the channel number
	// TODO: Create a function to do this

	return &client{
		// TODO: Add TCP connection
		send:    make(chan []byte, socketBufferSize),
		channel: &channel,
	}
}

// Handle handles events such as messages being sent
func Handle(c *client) {
	for {
		select {
		case message := <-c.send:
			// TODO: Send this to the server
			log.Printf("Sent:\n%v", hex.Dump(message))
		}
	}
}

func setUpServerConnection(c *client) {
	// Load client key pair
	cert, err := tls.LoadX509KeyPair(config.certs+"/client.pem", config.certs+"/client.key")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to load client key pair")
		os.Exit(1)
	}

	// Read the server certificate
	serverCert, err := ioutil.ReadFile(config.certs + "/server.pem")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to read server certificate file")
		os.Exit(1)
	}
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(serverCert)

	// Add both certificates to the TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
	}

	// Create a new connection
	conn, err := tls.Dial("tcp", "127.0.0.1:443", tlsConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial failed:", err.Error())
		os.Exit(1)
	}
	c.conn = conn

	// TODO: Authentication
	// Send the channel number to initialize the stream
	fmt.Fprintln(os.Stdout, "Sending stream initialization byte array.")
	_, err = c.conn.Write([]byte(""))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Writing stream init to server failed: ", err.Error())
		os.Exit(1)
	}
}
