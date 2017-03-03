package main

import (
	"os"
	"github.com/urfave/cli"
	"net"
	"sync"
	"strings"
	"strconv"
	log "github.com/Sirupsen/logrus"
	"time"
)

const (
	maxChannels      = 4               // maxChannels is the maximum number of channels supported
	timeout          = 5 * time.Second // timeout is the time before network operations timeout
	socketBufferSize = 1460            // socketBufferSize is the maximum buffer size for DVR data
)

// Config is a struct of all the configuration variables after user input is processed
type Config struct {
	source   *net.TCPAddr // source is the TCPAddr of the DVR
	dest     *net.TCPAddr // dest is the TCPAddr of the server
	user     string       // user is the username to authenticate with the DVR
	pass     string       // pass is the password to authenticate with the DVR
	key      string       // key is the passphrase to authenticate with the server
	channels []int        // channels is an array of currently used channels
	certs    string       // certs is the location to the folder storing client certificates
}

// Flags is a struct of the possible flags for CLI input
type Flags struct {
	user     string
	pass     string
	key      string
	source   string
	dest     string
	channels string
	certs    string
}

// Initialize global variables
var (
	config Config         // config stores the configuration values
	flags  Flags          // flags stores the flag values
	wg     sync.WaitGroup // wg stores the WaitGroup to prevent main.go from halting while routines are running
)

// main defines how the command line application works
func main() {
	// Create a new instance of urfave/cli
	app := cli.NewApp()

	// Each flag is saved in in the global flags variable
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "user", Value: "", Usage: "Username to authenticate with",
			Destination: &flags.user, EnvVar: "SWANN_USER", },
		cli.StringFlag{Name: "pass", Value: "", Usage: "Password to authenticate with",
			Destination: &flags.pass, EnvVar: "SWANN_PASS", },
		cli.StringFlag{Name: "source", Value: "", Usage: "The address of the DVR in the format host:port",
			Destination: &flags.source, EnvVar: "SWANN_SOURCE", },
		cli.StringFlag{Name: "dest", Value: "", Usage: "The address of the streaming server in the format host:port",
			Destination: &flags.dest, EnvVar: "SWANN_DEST", },
		cli.StringFlag{Name: "key", Value: "", Usage: "Passphrase to authenticate with the server",
			Destination: &flags.key, EnvVar: "SWANN_KEY"},
		cli.StringFlag{Name: "channels", Value: "", Usage: "Channel(s) to stream, delimited by commas",
			Destination: &flags.channels, EnvVar: "SWANN_CHANNELS", },
		cli.StringFlag{Name: "certs", Value: "", Usage: "Absolute file path to the certificate folder",
			Destination: &flags.certs, EnvVar: "SWANN_CERTS", },
	}

	app.Name = "swanntools-client"
	app.Usage = "client for kz/swanntools"
	app.Action = func(c *cli.Context) error {
		// Run the main application
		run()
		return nil
	}

	app.Run(os.Args)
}

// run handles the main running of the application
func run() {
	///////////////////////////////////////
	// 1. Validate and store flag values //
	///////////////////////////////////////

	// Assign empty Config struct to global config variable
	config = Config{}

	// Ensure that the command line flags are not empty
	if flags.user == "" || flags.pass == "" || flags.key == "" || flags.source == "" || flags.dest == "" ||
		flags.channels == "" || flags.certs == "" {
		log.Fatalln("You are missing one or more flags. Run --help for more details.")
	}

	// Add user, pass and key flags to config
	config.user = flags.user
	config.pass = flags.pass
	config.key = flags.key

	// Parse channel flag string (e.g., "1,3,4" -> ["1", "3", "4"])
	channelSlice := strings.Split(flags.channels, ",")

	// Ensure channels exist
	if len(channelSlice) == 0 {
		log.Fatalln("You must select a channel")
	}

	// Validate each flag
	for i, channel := range channelSlice {
		// Convert channel to integer
		intChannel, err := strconv.Atoi(channel)

		// Ensure maxChannels constraint is kept
		if i >= maxChannels {
			log.Fatalf("You cannot have greater than %d streams", maxChannels)
		} else if err != nil || intChannel > maxChannels {
			log.Fatalf("All channels need to be a number between 1 and %d", maxChannels)
		}

		// Ensure all channels are unique
		if intInSlice(&intChannel, &config.channels) {
			log.Fatalln("All channels need to be unique")
		}

		// Store channel number in config
		config.channels = append(config.channels, intChannel)
	}

	// Ensure certificates exist
	for _, file := range []string{"client.key", "client.pem", "server.pem"} {
		if _, err := os.Stat(flags.certs + "/" + file); err != nil {
			log.Fatalln("Unable to stat certificates: ", err.Error())
		}
	}

	// Store certificates in config
	config.certs = flags.certs

	//////////////////////////////////
	// 2. Resolve the TCP addresses //
	//////////////////////////////////

	// Resolve the source address
	tcpAddr, err := net.ResolveTCPAddr("tcp", flags.source)
	if err != nil {
		log.Fatalln("Resolving the source address failed: ", err.Error())
	}

	// Resolve the destination address
	tcpAddr, err = net.ResolveTCPAddr("tcp", flags.dest)
	if err != nil {
		log.Fatalln("Resolving the destination address failed: ", err.Error())
	}

	// Store addresses in config
	config.source = tcpAddr
	config.dest = tcpAddr

	////////////////////////////////////
	// 3. Retrieve the camera streams //
	////////////////////////////////////

	// Loop through each channel number
	for i := range config.channels {
		// Prevent main from exiting early before goroutines exit
		wg.Add(1)

		// Create a goroutine which streams channel to server
		go Stream{channel: &config.channels[i]}.StreamToServer()
	}

	// Wait for all goroutines to complete before exiting
	wg.Wait()
}
