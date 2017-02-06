package main

import (
	"os"
	"github.com/namsral/flag"
	"fmt"
	"encoding/hex"
)

const maxChannels = 4

// Config is a type of all the configuration variables after user input is processed
type Config struct {
	certs string
	key   string
}

// Initialize global variables
var (
	config Config
	// channelsInUse prevents the same channel from receiving multiple streams at once
	channelsInUse []int
)

func main() {
	config = Config{}

	// Retrieve the command line flags
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "SWANN", 0)
	certsInput := fs.String("certs", "", "Absolute file path to the certificate folder")
	keyInput := fs.String("key", "", "Passphrase to authenticate the client")
	fs.Parse(os.Args[1:])

	// Ensure that the command line flags are not empty
	if *certsInput == "" || *keyInput == "" {
		fs.PrintDefaults()
		os.Exit(1)
	}
	config.key = *keyInput
	println("Server listening for password:")
	println(hex.Dump([]byte(config.key)))

	// Ensure that the certificates exist at the location
	for _, file := range []string{"server.key", "server.pem"} {
		if _, err := os.Stat(*certsInput + "/" + file); err != nil {
			if err != nil {
				fmt.Fprintln(os.Stderr, "Unable to stat certificates: ", err.Error())
				os.Exit(1)
			}
		}
	}
	config.certs = *certsInput

	StartListener()
}
