package main

import (
	"os"
	"github.com/namsral/flag"
	"net"
	"sync"
	"strings"
	"strconv"
	"log"
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
		log.Fatalln("You are missing one or more flags. See above for the list of required flags.")
	}
	config.user = *userInput
	config.pass = *passInput
	config.key = *keyInput

	// Parse the channel input
	channelSlice := strings.Split(*channelInput, ",")
	for i, channel := range channelSlice {
		intChannel, err := strconv.Atoi(channel)
		if i >= maxChannels {
			log.Fatalf("You cannot have greater than %d streams", maxChannels)
		} else if err != nil || intChannel > maxChannels {
			log.Fatalf("All channels need to be a number between 1 and %d", maxChannels)
		}

		if intInSlice(&intChannel, &config.channels) {
			log.Fatalln("All channels need to be unique")
		}
		config.channels = append(config.channels, intChannel)
	}

	// Ensure that the certificates exist at the location
	for _, file := range []string{"client.key", "client.pem", "server.pem"} {
		if _, err := os.Stat(*certsInput + "/" + file); err != nil {
			log.Fatalln( "Unable to stat certificates: ", err.Error())
		}
	}
	config.certs = *certsInput

	// Resolve the TCP addresses
	tcpAddr, err := net.ResolveTCPAddr("tcp", *sourceInput)
	if err != nil {
		log.Fatalln("Resolving the source address failed: ", err.Error())
	}
	config.source = tcpAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", *destInput)
	if err != nil {
		log.Fatalln("Resolving the destination address failed: ", err.Error())
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
