package main

import (
	"crypto/tls"
	"net"
	"bufio"
	"strconv"
	"encoding/hex"
	"bytes"
	log "github.com/Sirupsen/logrus"
)

const (
	SuccessfulAuthString = "200"
	FailedAuthString     = "403"
	InvalidChannelString = "400"
	ChannelInUseString   = "409"
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
	log.Infof("Server ready and listening on: %s", config.bindAddr)

	for {
		// TODO: Use Mutexes to protect channels from simultaneous writes
		conn, err := listener.Accept()
		if err != nil {
			log.Warnln("An error occured when accepting a connection: ", err.Error())
			conn.Close()
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	var isAuthenticated bool = false
	var channel int
	var response string

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
		isAuthenticated, channel, response = parseAuthMessage(r)

		// Send appropriate response to the client
		if isAuthenticated {
			log.WithFields(log.Fields{
				"source": conn.RemoteAddr().String(), "channel": channel,
			}).Println("Auth success")
		} else {
			log.WithFields(log.Fields{
				"source": conn.RemoteAddr().String(), "channel": channel, "code": response,
			}).Warnln("Auth failure")
		}

		// Send the response to the client
		_, err := conn.Write([]byte(response))
		if err != nil {
			log.Infoln("Unable to write response to client: ", err.Error())
			break
		}
	}

	// Get the camera stream
	for {
		data := make([]byte, socketBufferSize)
		n, err := conn.Read(data)
		if err != nil {
			log.WithFields(log.Fields{
				"source": conn.RemoteAddr().String(), "channel": channel,
			}).Warnf("An error occurred while reading stream: %s", err.Error())
			break
		}

		// TODO: Do stuff with the camera stream!
		print(hex.Dump(data[:n]))
	}
}

func parseAuthMessage(r *bufio.Reader) (isAuthenticated bool, channelNum int, responseCode string) {
	var nilInt int

	// Parse channel and password
	msg, err := r.ReadString('\n')
	if err != nil {
		log.Warnln("Unable to retrieve authentication message: ", err.Error())
		return false, nilInt, FailedAuthString
	}
	channelInput := string(msg[0])
	// Ensure that line break is removed
	passwordInput := string(bytes.Trim([]byte(msg[1:]), "\x0a"))

	// Validate length, accounting for the line break
	if len(msg) < 3 {
		log.Warnln("Authentication failed due to invalid authentication message length")
		return false, nilInt, FailedAuthString
	}

	// Validate channel
	intChannel, err := strconv.Atoi(channelInput)
	if len(channelsInUse) >= maxChannels {
		log.Warnln("You cannot have greater than %d streams", maxChannels)
		return false, nilInt, InvalidChannelString
	} else if err != nil || intChannel > maxChannels {
		log.Warnln("All channels need to be a number between 1 and %d", maxChannels)
		return false, nilInt, InvalidChannelString
	} else if intInSlice(&intChannel, &channelsInUse) {
		log.Warnln("The channel %d is currently receiving a stream", intChannel)
		return false, nilInt, ChannelInUseString
	}

	// Validate password
	if passwordInput != config.key {
		return false, nilInt, FailedAuthString
	}

	return true, intChannel, SuccessfulAuthString
}
