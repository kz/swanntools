package main

import (
	"net"
	"fmt"
	"encoding/hex"
	"os"
	"bytes"
)

// Hex values of the byte arrays required in string form
const (
	initStreamValues     = "00000000000000000000010000000300000000000000000000006800000001000000100000000N0000000100000000UUUUUUUUUU000000000000000000000000000001000000000000010124000000PPPPPPPPPPPP00009cc9c805000000000400010004000000a8c9c80500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	successfulAuthValues = "1000000000000000"
	failedAuthValues     = "0800000004000000"
)

// stream represents the DVR camera stream
type stream struct {
	// conn is the TCP connection to the DVR
	conn *net.TCPConn
	// channel is the channel number of the stream
	channel *int
}

// setUpStreamConnection creates a new TCP connection
func setUpStreamConnection(s *stream) {
	// Attempt to close any existing TCP connections
	if s.conn != nil {
		s.conn.Close()
	}

	// Set up a new connection
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial failed:", err.Error())
		os.Exit(1)
	}
	s.conn = conn

	// Send the stream initialization byte array
	fmt.Fprintln(os.Stdout, "Sending stream initialization byte array.")
	_, err = s.conn.Write(initStreamBytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Writing stream init to server failed: ", err.Error())
		os.Exit(1)
	}

	// Check authentication response
	data := make([]byte, 8)
	_, err = s.conn.Read(data)
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

	// Add the connection to the stream
	s.conn = conn
}

// initStreamBytes returns the byte array required to initialize a stream
func initStreamBytes() []byte {
	hexValues := initStreamValues

	channelStartPos := 77
	channelEndPos := 78
	hexValues = hexValues[:channelStartPos] + fmt.Sprintf("%d", int(channel)) + hexValues[channelEndPos:]

	for i, v := range user {
		startPos := 94 + 2*i
		endPos := 96 + 2*i
		hexValues = hexValues[:startPos] + fmt.Sprintf("%x", int(v)) + hexValues[endPos:]
	}

	for i, v := range pass {
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
func StreamToServer(c *client, channel *int) {
	s := &stream{
		channel: channel,
	}
	setUpStreamConnection(s)

	defer s.conn.Close()

	// Start the handler to receive messages
	go Handle(c)

	// Get the main camera stream
	for {
		data := make([]byte, socketBufferSize)
		n, err := s.conn.Read(data)
		if err != nil {
			panic(err)
		}

		c.send <- data[:n]
	}
}

func StreamToStdout(channel *int) {
	s := &stream{
		channel: channel,
	}
	setUpStreamConnection(s)

	defer s.conn.Close()

	for {
		data := make([]byte, socketBufferSize)
		n, err := s.conn.Read(data)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(os.Stdout, "%s", data[:n])
	}
}
