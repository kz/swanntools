package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"log"
	"net"
	"bufio"
)

func StartListener() {
	// Load server key pair
	cert, err := tls.LoadX509KeyPair(config.certs+"/server.pem", config.certs+"/server.key")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to load server key pair")
		os.Exit(1)
	}
	// Add certificate to TLS config
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// TODO: Listen on flag
	listener, err := tls.Listen("tcp", "127.0.0.1:4000", tlsConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to start TLS listener")
		os.Exit(1)
	}

	for {
		// TODO: Use Mutexes to protect channels from simultaneous writes
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "An error occured when accepting a connection: ", err.Error())
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)

	isAuthenticated := false
	var channel int

	for {
		// TODO: Switch to handle authentication
		// If unauthenticated:
			// Parse channel and password
			// Create a stream (new stream.go with channel) if authentication successful
		// Otherwise, the message will be a stream
			// Forward the data to channel

		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		println(msg)
	}
}
