package main

import (
	"os"
	"github.com/namsral/flag"
	"fmt"
	"net"
	"math"
	"encoding/hex"
	"bytes"
	"log"
)

// Hex values of the byte arrays required in string form
const (
	InitStreamValues     = "00000000000000000000010000000300000000000000000000006800000001000000100000000N0000000100000000UUUUUUUUUU000000000000000000000000000001000000000000010124000000PPPPPPPPPPPP00009cc9c805000000000400010004000000a8c9c80500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	SuccessfulAuthValues = "1000000000000000"
	FailedAuthValues     = "0800000004000000"
)

// Initialize flag variables
var dest string
var user string
var pass string
var channel int

func main() {
	// Retrieve the command line flags
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "SWANN", 0)
	fs.StringVar(&dest, "dest", "", "The destination of the DVR in the format host:port")
	fs.StringVar(&user, "user", "", "Username to authenticate with")
	fs.StringVar(&pass, "pass", "", "Password to authenticate with")
	fs.IntVar(&channel, "channel", 0, "Channel to stream, either 1, 2, 3 or 4")
	fs.Parse(os.Args[1:])

	// Ensure that the command line flags are valid
	if dest == "" || user == "" || pass == "" || !(channel == 1 || channel == 2 || channel == 3 || channel == 4) {
		fs.PrintDefaults()
		os.Exit(1)
	}
	// Convert channel from 1, 2, 3, 4 to 1, 2, 4, 8 respectively
	channel = int(math.Exp2(float64(channel)))

	// Retrieve the camera stream
	cameraStream()
}

func initStreamBytes() []byte {
	hexValues := InitStreamValues

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

func cameraStream() {
	// Get the TCP address
	fmt.Fprintln(os.Stdout, "Establishing TCP connection.")
	tcpAddr, err := net.ResolveTCPAddr("tcp", dest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}

	// Set up the stream connection
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial failed:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// Send the stream initialization byte array
	fmt.Fprintln(os.Stdout, "Sending stream initialization byte array.")
	_, err = conn.Write(initStreamBytes())
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
	successfulAuthBytes, _ := hex.DecodeString(SuccessfulAuthValues)
	failedAuthBytes, _ := hex.DecodeString(FailedAuthValues)
	if bytes.Equal(data, successfulAuthBytes) {
		fmt.Fprintln(os.Stdout, "Successfully logged in!")
	} else if bytes.Equal(data, failedAuthBytes) {
		fmt.Fprintln(os.Stderr, "Authentication failed due to invalid credentials.")
		os.Exit(1)
	} else {
		fmt.Fprintln(os.Stderr, "Authentication failed due to unknown reason.")
		os.Exit(1)
	}

	// Get the main camera stream
	buf := &bytes.Buffer{}
	for {
		data := make([]byte, 1460)
		n, err := conn.Read(data)
		if err != nil {
			panic(err)
		}
		buf.Write(data[:n])
		log.Printf("Sent:\n%v", hex.Dump(data[:n]))
	}

}
