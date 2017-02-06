package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"net"
	"bufio"
	"strconv"
	"encoding/hex"
	"bytes"
)

const (
	SuccessfulAuthString = "200"
	FailedAuthString     = "403"
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

	println("Server ready")
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
		// TODO: Handle authentication
		if isAuthenticated {
			// Forward the data to channel
		} else {
			// Attempt authentication
			isAuthenticated, channel = parseAuthMessage(r)

			// Send appropriate response to the client
			var responseString string
			if isAuthenticated {
				responseString = SuccessfulAuthString
			} else {
				responseString = FailedAuthString
			}

			// Send the response to the client
			_, err := conn.Write([]byte(responseString))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to write response to client", err.Error())
				break
			}

			// Append the channel to slice of channels in use
			channelsInUse = append(channelsInUse, channel)
			// TODO: Add code to remove channel from channelsInUse
		}
	}
}

func parseAuthMessage(r *bufio.Reader) (bool, int) {
	var nilInt int

	// Parse channel and password
	msg, err := r.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to retrieve authentication message")
		return false, nilInt
	}
	channelInput := string(msg[0])
	// Ensure that line break is removed
	passwordInput := string(bytes.Trim([]byte(msg[1:]), "\x0a"))

	// Validate length, accounting for the line break
	if len(msg) < 3 {
		fmt.Fprintln(os.Stderr, "Authentication failed due to invalid authentication message length")
		return false, nilInt
	}

	// Validate channel
	intChannel, err := strconv.Atoi(channelInput)
	if len(channelsInUse) >= maxChannels {
		fmt.Fprintln(os.Stderr, "You cannot have greater than %d streams", maxChannels)
		return false, nilInt
	} else if err != nil || intChannel > maxChannels {
		fmt.Fprintln(os.Stderr, "All channels need to be a number between 1 and %d", maxChannels)
		return false, nilInt
	} else if intInSlice(&intChannel, &channelsInUse) {
		fmt.Fprintln(os.Stderr, "The channel %d is currently receiving a stream", intChannel)
		return false, nilInt
	}

	// Validate password
	println("Received password:")
	println(hex.Dump([]byte(passwordInput)))
	if passwordInput != config.key {
		fmt.Fprintln(os.Stderr, "Incorrect password")
		return false, nilInt
	}

	return true, intChannel
}
