package main

import (
	"net"
	"fmt"
	"encoding/hex"
	"bytes"
	log "github.com/Sirupsen/logrus"
	"math"
	"github.com/jpillora/backoff"
	"time"
)

// Hex values of the byte arrays required in string form
const (
	initStreamValues     = "00000000000000000000010000000300000000000000000000006800000001000000100000000N0000000100000000UUUUUUUUUU000000000000000000000000000001000000000000010124000000PPPPPPPPPPPP00009cc9c805000000000400010004000000a8c9c80500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	successfulAuthValues = "1000000000000000"
	failedAuthValues     = "0800000004000000"
)

// Stream is a struct handling streaming from the DVR
type Stream struct {
	channel   *int    // channel is a pointer to the DVR channel
	initBytes *[]byte // initBytes is a pointer to the byte array required to initialize a DVR stream
}

// newStreamConnection creates and sets up a new TCP connection
func (s *Stream) newStreamConnection() *net.Conn {
	// Generate the initBytes if it does not exist
	if s.initBytes == nil {
		s.generateInitBytes()
	}

	log.Infoln("Establishing connection and authenticating with the DVR...")

	// Add a backoff algorithm to handle network failures
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond, // Wait a minimum of 100 milliseconds
		Max:    30 * time.Second,       // Wait a maximum of 30 seconds
		Factor: 2,                      // Increase the wait factor by two each failure
		Jitter: false,                  // Disable jitter
	}

	// Create a local net.Conn variable
	var conn net.Conn

	// Use a ;; loop to handle network failure and backoff
	for {
		var err error

		// Attempt to dial the DVR with a timeout
		conn, err = net.DialTimeout("tcp", config.source.String(), timeout)
		if err != nil {
			log.Warnln("Dialing the DVR failed:", err.Error())
			// Increment the backoff duration
			d := b.Duration()
			// Wait for the backoff duration
			log.Infof("Retrying in %s...", d)
			time.Sleep(d)
			// Retry by restarting the loop
			continue
		}

		// Update the connection deadline with a new timeout
		conn.SetDeadline(time.Now().Add(timeout))

		// Send the stream initialization byte array to the DVR
		_, err = conn.Write(*s.initBytes)
		if err != nil {
			log.Warnln("Writing stream init to DVR failed: ", err.Error())
			// Close the connection as it is no longer untouched
			conn.Close()
			// Increment the backoff duration
			d := b.Duration()
			// Wait for the backoff duration
			log.Infof("Retrying in %s...", d)
			time.Sleep(d)
			// Retry by restarting the loop
			continue
		}

		// Reset the backoff and end the loop due to successful connection
		b.Reset()
		break
	}

	// Read the authentication response from the DVR
	data := make([]byte, 8)
	_, err := conn.Read(data)
	if err != nil {
		conn.Close()
		log.Fatalln("Unable to read DVR authentication response: ", err.Error())
	}

	// Check if DVR has authenticated the user
	successfulAuthBytes, _ := hex.DecodeString(successfulAuthValues)
	failedAuthBytes, _ := hex.DecodeString(failedAuthValues)
	if bytes.Equal(data, successfulAuthBytes) {
		log.Infoln("DVR authentication successful. Passing stream to client.")
	} else if bytes.Equal(data, failedAuthBytes) {
		conn.Close()
		log.Fatalln("DVR authentication failed due to invalid credentials.")
	} else {
		conn.Close()
		log.Fatalln("DVR authentication failed due to unknown reason.")
	}

	// Return the stream
	return &conn
}

// generateInitBytes returns the byte array required to initialize a stream
func (s *Stream) generateInitBytes() {
	hexValues := initStreamValues

	channelStartPos := 77
	channelEndPos := 78

	// Convert channel from 1, 2, 3, 4 to 1, 2, 4, 8 respectively
	parsedChannel := int(math.Exp2(float64(*s.channel - 1)))
	hexValues = hexValues[:channelStartPos] + fmt.Sprintf("%d", parsedChannel) + hexValues[channelEndPos:]

	// Parse username
	for i, v := range config.user {
		startPos := 94 + 2*i
		endPos := 96 + 2*i
		hexValues = hexValues[:startPos] + fmt.Sprintf("%x", int(v)) + hexValues[endPos:]
	}

	// Parse password
	for i, v := range config.pass {
		startPos := 158 + 2*i
		endPos := 160 + 2*i
		hexValues = hexValues[:startPos] + fmt.Sprintf("%x", int(v)) + hexValues[endPos:]
	}

	// Decode the hex string into a byte array
	byteArray, err := hex.DecodeString(hexValues)
	if err != nil {
		log.Fatalln("Unable to decode stream initialization values to byte array: ", err.Error())
	}

	s.initBytes = &byteArray
}

// StreamToServer streams the video to the server
func (s *Stream) StreamToServer() {
	// Remove a WaitGroup entry once stream halts so main can exit
	defer wg.Done()

	// Create a new stream connection
	conn := *s.newStreamConnection()
	defer conn.Close()

	// Create a client and handler to receive messages
	c := Client(s.channel)

	// Run the client handler in a goroutine
	go c.Handle()

	// Get the main camera stream and send it to the client handler
	for {
		// Create a byte array
		data := make([]byte, socketBufferSize)

		// Update the DVR conn timeout
		conn.SetDeadline(time.Now().Add(timeout))

		// Read from the DVR
		n, err := conn.Read(data)
		if err != nil {
			log.Warnln("Error occurred while reading from DVR stream connection: ", err.Error())
			// Close the connection
			conn.Close()
			// Reattempt the connection
			conn = *s.newStreamConnection()
			// Loop again and listen for more data
			continue
		}

		// Send the data to the c.send chan for handling by the client handler
		c.send <- data[:n]
	}
}
