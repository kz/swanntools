package main

import (
	"os"
	"github.com/namsral/flag"
	"fmt"
)

const maxChannels = 4

// Config is a type of all the configuration variables after user input is processed
type Config struct {
	certs string
}

// Initialize global variables
var config Config

func main() {
	config = Config{}

	// Retrieve the command line flags
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "SWANN", 0)
	certsInput := fs.String("certs", "", "Absolute file path to the certificate folder")
	fs.Parse(os.Args[1:])

	// Ensure that the command line flags are not empty
	if *certsInput == "" {
		fs.PrintDefaults()
		os.Exit(1)
	}

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
