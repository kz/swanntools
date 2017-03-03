package main

import (
	"crypto/x509"
	"crypto/tls"
	"io/ioutil"
	"strconv"
	"github.com/jpillora/backoff"
	"time"
	log "github.com/Sirupsen/logrus"
	"net"
)

// Server responses
const (
	SuccessfulClientAuthString = "200"
	FailedClientAuthString     = "403"
	InvalidChannelString       = "400"
	ChannelInUseString         = "409"
)

// client represents the local machine sending the DVR stream
type client struct {
	conn    *tls.Conn   // conn is the TCP (w/ TLS) connection to the server
	send    chan []byte // send is the channel on which messages are sent
	channel *int        // channel is the channel number of the stream
}

// Client creates a new client struct
func Client(channel *int) *client {
	c := &client{channel: channel}
	c.conn = c.newServerConnection()
	c.send = make(chan []byte, socketBufferSize)
	return c
}

// Handle handles events such as messages being sent
func (c *client) Handle() {
	for {
		select {
		// Handles sending of video stream data to the server
		case message := <-c.send:
			// Update the deadline for the server connection
			c.conn.SetDeadline(time.Now().Add(timeout))
			// Write the data to the server
			_, err := c.conn.Write(message)
			if err != nil {
				log.Warnln("Error occurred while writing to server: ", err.Error())
				log.Infoln("Attempting to reestablish connection...")
				// Close the connection
				c.conn.Close()
				// Reattempt the connection
				c.conn = c.newServerConnection()
				// Loop again
				continue
			}
		}
	}
}

// newServerConnection creates a new TLS connection with the server
func (c *client) newServerConnection() *tls.Conn {
	////////////////////////////
	// 1. Handle certificates //
	////////////////////////////

	// Load client key pair for TLS connection
	clientCerts, err := tls.LoadX509KeyPair(config.certs+"/client.pem", config.certs+"/client.key")
	if err != nil {
		log.Fatalln("Unable to load client key pair")
	}

	// Read the server certificate
	serverCert, err := ioutil.ReadFile(config.certs + "/server.pem")
	if err != nil {
		log.Fatalln("Unable to read server certificate file")
	}
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(serverCert)

	// Add both certificates to the TLS config
	tlsConfig := &tls.Config{
		// TODO: Generate CA-signed certificates instead
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientCerts},
		RootCAs:            roots,
	}

	//////////////////////////////
	// 2. Connect to the server //
	//////////////////////////////

	log.Infoln("Establishing connection and authenticating with server...")

	// Add a backoff/retry algorithm
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Minute,
		Factor: 2,
		Jitter: false,
	}

	// Create a local conn variable
	var conn *tls.Conn
	// Create a variable to store the server authentication response
	var authResponse []byte

	// Use a ;; loop to handle network failure and backoff
	for {
		// Create a new connection with a timeout
		dialer := &net.Dialer{Timeout: timeout}
		conn, err = tls.DialWithDialer(dialer, "tcp", config.dest.String(), tlsConfig)
		if err != nil {
			// Increment the backoff duration
			d := b.Duration()
			log.Warnln("Unable to dial the server: ", err.Error())
			// Wait for the backoff duration
			log.Infof("Retrying in %s...", d)
			time.Sleep(d)
			// Retry by restarting the loop
			continue
		}

		// Update the connection deadline with a new timeout
		conn.SetDeadline(time.Now().Add(timeout))

		// Send the channel number along with login details
		_, err = conn.Write([]byte(strconv.Itoa(*c.channel) + config.key + "\n"))
		if err != nil {
			// Close the connection as it is no longer untouched
			conn.Close()
			// Increment the backoff duration
			d := b.Duration()
			log.Warnln("Writing stream init to server failed: ", err.Error())
			// Wait for the backoff duration
			log.Infof("Retrying in %s...", d)
			time.Sleep(d)
			// Retry by restarting the loop
			continue
		}

		// Create a byte array to store the server response
		authResponse = make([]byte, 3)
		// Read the server response
		_, err = conn.Read(authResponse)
		if err != nil {
			// Close the connection as it is no longer untouched
			conn.Close()
			// Increment the backoff duration
			d := b.Duration()
			log.Warnln("Unable to read the server authentication response: ", err.Error())
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

	/////////////////////////////////////
	// 3. Authenticate with the server //
	/////////////////////////////////////

	// Check authResponse with auth strings
	if string(authResponse) == SuccessfulClientAuthString {
		log.Infoln("Successfully authenticated with the server. Passing stream to server.")
	} else if string(authResponse) == FailedClientAuthString {
		conn.Close()
		log.Fatalln("Authentication failed due to invalid credentials.")
	} else if string(authResponse) == InvalidChannelString {
		conn.Close()
		log.Fatalln("Authentication failed due to invalid channel provided.")
	} else if string(authResponse) == ChannelInUseString {
		conn.Close()
		log.Fatalln("Unable to connect as channel is in use.")
	} else {
		conn.Close()
		log.Fatalln("Authentication failed due to unknown reason.")
	}

	return conn
}
