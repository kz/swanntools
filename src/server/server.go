package main

import (
	"crypto/tls"
	"net"
	"bufio"
	"strconv"
	"encoding/hex"
	"bytes"
	"log"
)

const (
	SuccessfulAuthString = "200"
	FailedAuthString     = "403"
	socketBufferSize     = 1460
)

func StartListener() {
	// Load server key pair
	cert, err := tls.LoadX509KeyPair(config.certs+"/server.pem", config.certs+"/server.key")
	if err != nil {
		log.Fatalln("Unable to load server key pair: ", err.Error())
	}

	// Add certificate to TLS config
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	listener, err := tls.Listen("tcp", config.bindAddr.String(), tlsConfig)
	if err != nil {
		log.Fatalln("Unable to start TLS listener: ", err.Error())
	}

	// Print debugging messages
	log.Printf("Server listening on: %s\n", config.bindAddr)
	log.Println("Server listening for password:")
	log.Println(hex.Dump([]byte(config.key)))
	log.Println("Server ready")

	for {
		// TODO: Use Mutexes to protect channels from simultaneous writes
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("An error occured when accepting a connection: ", err.Error())
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	isAuthenticated := false
	var channel int

	defer func(channel *int) {
		conn.Close()
		if channel != nil {
			// Remove channel from channelsInUse if appropriate
			if pos, isPresent := intPositionInSlice(channel, &channelsInUse); isPresent {
				channelsInUse = append(channelsInUse[:pos], channelsInUse[pos+1:]...)
			}
		}
	}(&channel)

	// Handle authentication
	r := bufio.NewReader(conn)
	for {
		if isAuthenticated {
			// Append the channel to slice of channels in use
			channelsInUse = append(channelsInUse, channel)
			break
		}

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
			log.Panicln("Unable to write response to client: ", err.Error())
		}

		log.Println("Successfully authenticated!")
	}

	// Get the camera stream
	for {
		data := make([]byte, socketBufferSize)
		n, err := conn.Read(data)
		if err != nil {
			panic(err)
		}

		// TODO: Do stuff with the camera stream!
		print(hex.Dump(data[:n]))
	}
}

func parseAuthMessage(r *bufio.Reader) (bool, int) {
	var nilInt int

	// Parse channel and password
	msg, err := r.ReadString('\n')
	if err != nil {
		log.Println("Unable to retrieve authentication message: ", err.Error())
		return false, nilInt
	}
	channelInput := string(msg[0])
	// Ensure that line break is removed
	passwordInput := string(bytes.Trim([]byte(msg[1:]), "\x0a"))

	// Validate length, accounting for the line break
	if len(msg) < 3 {
		log.Println("Authentication failed due to invalid authentication message length")
		return false, nilInt
	}

	// Validate channel
	intChannel, err := strconv.Atoi(channelInput)
	if len(channelsInUse) >= maxChannels {
		log.Printf("You cannot have greater than %d streams\n", maxChannels)
		return false, nilInt
	} else if err != nil || intChannel > maxChannels {
		log.Printf("All channels need to be a number between 1 and %d\n", maxChannels)
		return false, nilInt
	} else if intInSlice(&intChannel, &channelsInUse) {
		log.Printf("The channel %d is currently receiving a stream\n", intChannel)
		return false, nilInt
	}

	// Validate password
	log.Println("Received password:")
	log.Println(hex.Dump([]byte(passwordInput)))
	if passwordInput != config.key {
		log.Println("Incorrect password")
		return false, nilInt
	}

	return true, intChannel
}
