package main

import (
	"net"
	"fmt"
	"encoding/hex"
	"os"
	"bytes"
	"log"
	"strconv"
	"math"
)

// Hex values of the byte arrays required in string form
const (
	initStreamValues     = "00000000000000000000010000000300000000000000000000006800000001000000100000000N0000000100000000UUUUUUUUUU000000000000000000000000000001000000000000010124000000PPPPPPPPPPPP00009cc9c805000000000400010004000000a8c9c80500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	successfulAuthValues = "1000000000000000"
	failedAuthValues     = "0800000004000000"
)

// newStreamConnection creates and sets up a new TCP connection
func newStreamConnection(channel *int) *net.TCPConn {
	// Set up a new connection
	conn, err := net.DialTCP("tcp", nil, config.source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial failed:", err.Error())
		os.Exit(1)
	}

	// Send the stream initialization byte array
	fmt.Fprintln(os.Stdout, "Sending stream initialization byte array.")
	_, err = conn.Write(initStreamBytes(channel))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Writing stream init to server failed: ", err.Error())
		os.Exit(1)
	}

	// Check authentication response
	data := make([]byte, 8)
	_, err = conn.Read(data)
	if err != nil {
		panic(err)
	}

	// Check if authenticated
	successfulAuthBytes, _ := hex.DecodeString(successfulAuthValues)
	failedAuthBytes, _ := hex.DecodeString(failedAuthValues)
	if bytes.Equal(data, successfulAuthBytes) {
		fmt.Fprintln(os.Stdout, "Successfully logged in!")
	} else if bytes.Equal(data, failedAuthBytes) {
		fmt.Fprintln(os.Stderr, "Authentication failed due to invalid credentials.")
		os.Exit(1)
	} else {
		fmt.Fprintln(os.Stderr, "Authentication failed due to unknown reason.")
		os.Exit(1)
	}

	// Return the stream
	return conn
}

// initStreamBytes returns the byte array required to initialize a stream
func initStreamBytes(channel *int) []byte {
	hexValues := initStreamValues

	channelStartPos := 77
	channelEndPos := 78
	// Convert channel from 1, 2, 3, 4 to 1, 2, 4, 8 respectively
	parsedChannel := int(math.Exp2(float64(*channel - 1)))
	hexValues = hexValues[:channelStartPos] + fmt.Sprintf("%d", parsedChannel) + hexValues[channelEndPos:]

	for i, v := range config.user {
		startPos := 94 + 2*i
		endPos := 96 + 2*i
		hexValues = hexValues[:startPos] + fmt.Sprintf("%x", int(v)) + hexValues[endPos:]
	}

	for i, v := range config.pass {
		startPos := 158 + 2*i
		endPos := 160 + 2*i
		hexValues = hexValues[:startPos] + fmt.Sprintf("%x", int(v)) + hexValues[endPos:]
	}

	byteArray, err := hex.DecodeString(hexValues)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to decode stream initialization values to byte array: ", err.Error())
		os.Exit(1)
	}
	return byteArray
}

// StreamToServer streams the video to the server
func StreamToServer(channel *int) {
	defer wg.Done()

	conn := newStreamConnection(channel)
	defer conn.Close()

	// Create a client
	c := Client(channel)

	// Start the client handler to receive messages
	go Handle(c)

	// Get the main camera stream
	for {
		data := make([]byte, socketBufferSize)
		n, err := conn.Read(data)
		if err != nil {
			panic(err)
		}

		c.send <- data[:n]
	}
}

func StreamToStdout(channel *int) {
	defer wg.Done()

	conn := newStreamConnection(channel)
	defer conn.Close()

	for {
		data := make([]byte, socketBufferSize)
		n, err := conn.Read(data)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(os.Stdout, "%s", data[:n])
	}
}

func StreamToFile(channel *int) {
	defer wg.Done()

	// Attempt to create output file
	fileName := "swann_" + strconv.Itoa(*channel)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// Open file for writing
	file, err = os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	print("Writing to: " + fileName)
	defer file.Close()

	conn := newStreamConnection(channel)
	defer conn.Close()

	for {
		data := make([]byte, socketBufferSize)
		n, err := conn.Read(data)
		if err != nil {
			panic(err)
		}

		file.Write(data[:n])
	}
}
