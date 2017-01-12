package main

import (
	"flag"
	"os"
	"encoding/hex"
	"strconv"
	"fmt"
	"net"
	"bytes"
)

// Declare values of the byte arrays required in string form
const (
	IntentValues          = "00000000000000000000010000000aXX000000292300000000001c010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	IntentResponseValues  = "000000010000000aXX000000292300000000001c010000000100961200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	AuthValues            = "000000000000000000000100000019YY0000000000000000000054000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	SuccessfulLoginValues = "0800000002000000"
	FailedLoginValues     = "08000000FFFFFFFF"
)

// Initialize flag variables
var dest string
var user string
var pass string

func init() {
	// Retrieve the command line flags
	flag.StringVar(&dest, "dest", "", "The destination of the DVR in the format host:port")
	flag.StringVar(&user, "user", "", "Username to authenticate with")
	flag.StringVar(&pass, "pass", "", "Password to authenticate with")
}

func getIntentMessage(intentValue string) []byte {
	hexValues := IntentValues[:30] + intentValue + IntentValues[32:]
	byteArray, err := hex.DecodeString(hexValues)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to decode intent message to byte array: ", err.Error())
		os.Exit(1)
	}
	return byteArray
}

func getIntentResponseMessage(intentValue string) []byte {
	hexValues := IntentResponseValues[:16] + intentValue + IntentResponseValues[18:]
	byteArray, err := hex.DecodeString(hexValues)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to decode intent response message to byte array: ", err.Error())
		os.Exit(1)
	}
	return byteArray
}

func getLoginMessage(user string, pass string, intentValue string) []byte {
	// Add incremented intent value to hex values
	parsedIntentValue, err := strconv.ParseInt(intentValue, 16, 8)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to parse intent value to inty: ", err.Error())
		os.Exit(1)
	}
	incIntentHexValue := fmt.Sprintf("%x", parsedIntentValue+1)
	hexValues := AuthValues[:30] + incIntentHexValue + AuthValues[32:]

	// Add username to hex values
	for i, v := range user {
		startPos := 54 + 2*i
		endPos := 56 + 2*i
		hexValues = hexValues[:startPos] + fmt.Sprintf("%x", int(v)) + hexValues[endPos:]
	}

	// Add password to hex values
	for i, v := range pass {
		startPos := 118 + 2*i
		endPos := 120 + 2*i
		hexValues = hexValues[:startPos] + fmt.Sprintf("%x", int(v)) + hexValues[endPos:]
	}

	byteArray, err := hex.DecodeString(hexValues)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to decode login message to byte array: ", err.Error())
		os.Exit(1)
	}
	return byteArray

}

// Reference: https://gist.github.com/iwanbk/2295233
func sendIntent(intentValue string, tcpAddr *net.TCPAddr) {
	intentMessage := getIntentMessage(intentValue)
	intentResponseMessage := getIntentResponseMessage(intentValue)

	// Set up the intent connection
	intentConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial failed:", err.Error())
		os.Exit(1)
	}
	defer intentConn.Close()

	// Send the intent message
	fmt.Fprintln(os.Stdout, "Sending intent message.")
	_, err = intentConn.Write(intentMessage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Writing intent message to server failed: ", err.Error())
		os.Exit(1)
	}

	// Receive the intent response message
	fmt.Fprintln(os.Stdout, "Receiving intent response message.")
	intentReply := make([]byte, 500)
	_, err = intentConn.Read(intentReply)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection read of intent message response failed:", err.Error())
		os.Exit(1)
	}

	// Check whether the intent message response is as expected
	fmt.Fprintln(os.Stdout, "Checking intent response message.")
	if !bytes.Equal(intentReply, intentResponseMessage) {
		fmt.Fprintln(os.Stderr, "Intent message response not as expected: ", string(intentReply))
		os.Exit(1)
	}
}

func sendLogin(intentValue string, tcpAddr *net.TCPAddr) {
	loginMessage := getLoginMessage(user, pass, intentValue)
	successfulLoginMessage, _ := hex.DecodeString(SuccessfulLoginValues)
	failedLoginMessage, _ := hex.DecodeString(FailedLoginValues)

	// Set up the login connection
	loginConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial failed:", err.Error())
		os.Exit(1)
	}
	defer loginConn.Close()

	// Send the login message
	fmt.Fprintln(os.Stdout, "Sending login message.")
	_, err = loginConn.Write(loginMessage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Writing intent message to server failed: ", err.Error())
		os.Exit(1)
	}

	// Receive the login response message
	fmt.Fprintln(os.Stdout, "Receiving login response message.")
	loginReply := make([]byte, 8)
	_, err = loginConn.Read(loginReply)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection read of intent message response failed:", err.Error())
		os.Exit(1)
	}

	// Check the login response message status
	fmt.Fprintln(os.Stdout, "Checking login response message.")
	if bytes.Equal(loginReply, successfulLoginMessage) {
		fmt.Fprintln(os.Stdout, "Successfully logged in!")
	} else if bytes.Equal(loginReply, failedLoginMessage) {
		fmt.Fprintln(os.Stderr, "Authentication failed due to invalid credentials.")
		os.Exit(1)
	} else {
		fmt.Fprintln(os.Stderr, "Authentication failed due to unknown reason.")
		os.Exit(1)
	}
}

func sendSettings(tcpAddr *net.TCPAddr) {
	// Send the DVR setup message request
	SetupMessageValues := "00000000000000000000010000000e1c0000000000000000000014000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	setupMessage, _ := hex.DecodeString(SetupMessageValues)
	setupConn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Dial failed:", err.Error())
		os.Exit(1)
	}
	defer setupConn.Close()

	// Send the setup message
	fmt.Fprintln(os.Stdout, "Sending setup message.")
	_, err = setupConn.Write(setupMessage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Writing setup message to server failed: ", err.Error())
		os.Exit(1)
	}

	// Receive the setup response message
	fmt.Fprintln(os.Stdout, "Receiving setup response message.")
	setupReply := make([]byte, 13000)
	setupRead, err := setupConn.Read(setupReply)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Connection read of setup message response failed:", err.Error())
		os.Exit(1)
	}

	println(string(setupRead))
}

func main() {
	// Take environment variables to have higher precedence than command line flags
	destEnv, userEnv, passEnv := os.Getenv("AUTH_DEST"), os.Getenv("AUTH_USER"), os.Getenv("AUTH_PASS")
	if destEnv != "" && userEnv != "" && passEnv != "" {
		dest, user, pass = destEnv, userEnv, passEnv
	}

	// Otherwise, ensure that the command line flags are not empty
	if dest == "" || user == "" || pass == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Get the TCP address
	fmt.Fprintln(os.Stdout, "Establishing TCP connection.")
	tcpAddr, err := net.ResolveTCPAddr("tcp", dest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}

	sendIntent("1a", tcpAddr)
	sendLogin("1a", tcpAddr)
	sendSettings(tcpAddr)
}
