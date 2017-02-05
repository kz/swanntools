package main

import (
	"os"
	"github.com/namsral/flag"
	"fmt"
	"net"
	"sync"
	"strings"
	"math"
	"strconv"
)

const maxChannels = 4

// Config is a type of all the configuration variables after user input is processed
type Config struct {
	source   *net.TCPAddr
	dest     *net.TCPAddr
	user     string
	pass     string
	key      string
	channels []int
	certs    string
}

// Initialize global variables
var (
	config Config
	wg     sync.WaitGroup
)

func main() {
	config = Config{}

	// Retrieve the command line flags
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "SWANN", 0)
	userInput := fs.String("user", "", "Username to authenticate with")
	passInput := fs.String("pass", "", "Password to authenticate with")
	keyInput := fs.String("key", "", "Passphrase to authenticate with the server")
	sourceInput := fs.String("source", "", "The address of the DVR in the format host:port")
	destInput := fs.String("dest", "", "The address of the streaming server in the format host:port")
	channelInput := fs.String("channels", "", "Channel(s) to stream, delimited by commas")
	certsInput := fs.String("certs", "", "Absolute file path to the certificate folder")

	fs.Parse(os.Args[1:])

	// Ensure that the command line flags are not empty
	if *userInput == "" || *passInput == "" || *keyInput == "" || *sourceInput == "" || *destInput == "" ||
		*channelInput == "" || *certsInput == "" {
		fs.PrintDefaults()
		os.Exit(1)
	}
	config.user = *userInput
	config.pass = *passInput
	config.key = *keyInput

	// Parse the channel input
	channelSlice := strings.Split(*channelInput, ",")
	for i, channel := range channelSlice {
		intChannel, err := strconv.Atoi(channel)
		if i >= maxChannels {
			fmt.Fprintln(os.Stderr, "You cannot have greater than %d streams", maxChannels)
			os.Exit(1)
		} else if err != nil || intChannel > maxChannels {
			fmt.Fprintln(os.Stderr, "All channels need to be a number between 1 and %d", maxChannels)
			os.Exit(1)
		}
		// Convert channel from 1, 2, 3, 4 to 1, 2, 4, 8 respectively
		parsedChannel := int(math.Exp2(float64(intChannel - 1)))
		if intInSlice(&parsedChannel, &config.channels) {
			fmt.Fprintln(os.Stderr, "All channels need to be unique", maxChannels)
			os.Exit(1)
		}
		config.channels = append(config.channels, parsedChannel)
	}

	// Ensure that the certificates exist at the location
	for _, file := range []string{"client.key", "client.pem", "server.key", "server.pem"} {
		if _, err := os.Stat(*certsInput + "/" + file); err != nil {
			fmt.Fprintln(os.Stderr, "Unable to stat certificates: ", err.Error())
			os.Exit(1)
		}
	}
	config.certs = *certsInput

	// Resolve the TCP addresses
	tcpAddr, err := net.ResolveTCPAddr("tcp", *sourceInput)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}
	config.source = tcpAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", *destInput)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}
	config.dest = tcpAddr

	// Retrieve the camera streams
	for i := range config.channels {
		wg.Add(1)
		go StreamToServer(&config.channels[i])
	}

	wg.Wait()
}
