package main

import (
	"os"
	"github.com/urfave/cli"
	"net"
	log "github.com/Sirupsen/logrus"
)

const maxChannels = 4

// Config is a type of all the configuration variables after user input is processed
type Config struct {
	bindAddr *net.TCPAddr
	key      string
	certs    string
}

type Flags struct {
	bindAddr string
	key      string
	certs    string
}

// Initialize global variables
var (
	flags  Flags
	config Config
	// channelsInUse prevents the same channel from receiving multiple streams at once
	channelsInUse []int
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "bind", Value: "", Usage: "The address to listen on in the format host:port",
			Destination: &flags.bindAddr, EnvVar: "SWANN_BIND", },
		cli.StringFlag{Name: "key", Value: "", Usage: "Passphrase to authenticate the client",
			Destination: &flags.key, EnvVar: "SWANN_KEY"},
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
	if flags.certs == "" || flags.key == "" || flags.bindAddr == "" {
		log.Fatalln("You are missing one or more flags. Run --help for more details.")
	}
	config.key = flags.key

	// Ensure that the certificates exist at the location
	for _, file := range []string{"server.key", "server.pem"} {
		if _, err := os.Stat(flags.certs + "/" + file); err != nil {
			if err != nil {
				log.Fatalln("Unable to stat certificates: ", err.Error())
			}
		}
	}
	config.certs = flags.certs

	// Resolve the TCP address to bind to
	tcpAddr, err := net.ResolveTCPAddr("tcp", flags.bindAddr)
	if err != nil {
		log.Fatalln("Resolving the bind address failed: ", err.Error())
	}
	config.bindAddr = tcpAddr

	StartListener()
}
