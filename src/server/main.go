package main

import (
	"os"
	"github.com/urfave/cli"
	"net"
	log "github.com/Sirupsen/logrus"
)

const (
	maxChannels      = 4    // maxChannels is the maximum number of channels supported
	socketBufferSize = 1460 // socketBufferSize is the maximum buffer size for DVR data
)

// Config is a struct of all the configuration variables after user input is processed
type Config struct {
	bindAddr  *net.TCPAddr // bindAddr is the TCP address for the server to bind to
	key       string       // key is the passphrase to authenticate the client with
	certs     string       // certs is the file path to the server certificates
	consumers []Consumer   // consumer is a byte array of consumers which performs actions on the stream
}

// Flags is a struct of all flags after user input is processed
type Flags struct {
	bindAddr string
	key      string
	certs    string
	saveDisk string
}

// Initialize global variables
var (
	flags         Flags  // flags stores the CLI flags
	config        Config // config stores the configuration values
	channelsInUse []int  // channelsInUse prevents the same channel from receiving multiple streams at once
)

// main defines how the command line application works
func main() {
	// Create a new instance of urfave/cli
	app := cli.NewApp()

	// Each flag is saved in in the global flags variable
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "bind", Value: "", Usage: "The address to listen on in the format host:port",
			Destination: &flags.bindAddr, EnvVar: "SWANN_BIND", },
		cli.StringFlag{Name: "key", Value: "", Usage: "Passphrase to authenticate the client",
			Destination: &flags.key, EnvVar: "SWANN_KEY"},
		cli.StringFlag{Name: "certs", Value: "", Usage: "Absolute file path to the certificate folder",
			Destination: &flags.certs, EnvVar: "SWANN_CERTS", },
		cli.StringFlag{Name: "save-disk", Value: "", Usage: "File path to transcode and save the stream to",
			Destination: &flags.saveDisk, EnvVar: "SWANN_SAVE_DISK"},
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
	// Assign empty Config struct to global config variable
	config = Config{}

	// Ensure that the command line flags are not empty
	if flags.certs == "" || flags.key == "" || flags.bindAddr == "" {
		log.Fatalln("You are missing one or more flags. Run --help for more details.")
	}

	// Add key flag to config
	config.key = flags.key

	// Ensure that the certificates exist at the location
	for _, file := range []string{"server.key", "server.pem"} {
		if _, err := os.Stat(flags.certs + "/" + file); err != nil {
			log.Fatalln("Unable to stat certificates: ", err.Error())
		}
	}

	// Add certificate to config
	config.certs = flags.certs

	// If saveDisk is set, ensure that directory exists and start handler
	if flags.saveDisk != "" {
		// Check if directory exists
		if _, err := os.Stat(flags.saveDisk); err != nil {
			log.Fatalln("Unable to stat save disk folder: ", err.Error())
		}

		// Append a new consumer to config.consumers
		config.consumers = append(config.consumers, Consumer{
			Receiver:    make(chan []byte, socketBufferSize),
			HandlerType: SaveDiskHandlerType,
		})
	}

	// Resolve the TCP address to bind to
	tcpAddr, err := net.ResolveTCPAddr("tcp", flags.bindAddr)
	if err != nil {
		log.Fatalln("Resolving the bind address failed: ", err.Error())
	}

	// Add bindAddr to config
	config.bindAddr = tcpAddr

	// Start the server listener
	StartListener()
}
