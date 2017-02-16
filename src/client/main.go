package main

import (
	"os"
	"github.com/urfave/cli"
	"net"
	"sync"
	"strings"
	"strconv"
	log "github.com/Sirupsen/logrus"
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
	config Config
	flags  Flags
	wg     sync.WaitGroup
)

func main() {
	app := cli.NewApp()

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
		run()
		return nil
	}

	app.Run(os.Args)
}

func run() {
	config = Config{}

	// Ensure that the command line flags are not empty
	if flags.user == "" || flags.pass == "" || flags.key == "" || flags.source == "" || flags.dest == "" ||
		flags.channels == "" || flags.certs == "" {
		log.Fatalln("You are missing one or more flags. Run --help for more details.")
	}
	config.user = flags.user
	config.pass = flags.pass
	config.key = flags.key

	// Parse the channel input
	channelSlice := strings.Split(flags.channels, ",")
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
		if _, err := os.Stat(flags.certs + "/" + file); err != nil {
			log.Fatalln("Unable to stat certificates: ", err.Error())
		}
	}
	config.certs = flags.certs

	// Resolve the TCP addresses
	tcpAddr, err := net.ResolveTCPAddr("tcp", flags.source)
	if err != nil {
		log.Fatalln("Resolving the source address failed: ", err.Error())
	}
	config.source = tcpAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", flags.dest)
	if err != nil {
		log.Fatalln("Resolving the destination address failed: ", err.Error())
	}
	config.dest = tcpAddr

	// Retrieve the camera streams
	for i := range config.channels {
		wg.Add(1)
		go StreamToServer(&config.channels[i])
	}

	wg.Wait()
}
