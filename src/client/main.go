package main

import (
	"os"
	"github.com/namsral/flag"
	"fmt"
	"net"
	"math"
	"strings"
	"unicode"
)

const maxChannels = 4

// Initialize flag variables
var (
	addr     *net.TCPAddr
	dest     string
	user     string
	pass     string
	channels []int
)

func main() {
	// Retrieve the command line flags
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "SWANN", 0)
	fs.StringVar(&dest, "dest", "", "The destination of the DVR in the format host:port")
	fs.StringVar(&user, "user", "", "Username to authenticate with")
	fs.StringVar(&pass, "pass", "", "Password to authenticate with")
	channelInput := fs.String("channels", "", "Channel(s) to stream, delimited by commas")
	fs.Parse(os.Args[1:])

	// Ensure that the command line flags are not empty
	if dest == "" || user == "" || pass == "" || *channelInput == "" {
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Parse the channel input
	channelSlice := strings.Split(*channelInput, ",")
	for i, channel := range channelSlice {
		if i >= maxChannels {
			fmt.Fprintln(os.Stderr, "You cannot have greater than %d streams", maxChannels)
			os.Exit(1)
		} else if !unicode.IsDigit(rune(channel)) {
			fmt.Fprintln(os.Stderr, "All channels need to be a number between 1 and %d", maxChannels)
			os.Exit(1)
		}
		// Convert channel from 1, 2, 3, 4 to 1, 2, 4, 8 respectively
		parsedChannel := int(math.Exp2(float64(int(channel) - 1)))
		if intInSlice(&parsedChannel, &channels) {
			fmt.Fprintln(os.Stderr, "All channels need to be unique", maxChannels)
			os.Exit(1)
		}
		channels = append(channels, parsedChannel)
	}

	// Resolve the TCP address
	fmt.Fprintln(os.Stdout, "Resolving TCP address.")
	tcpAddr, err := net.ResolveTCPAddr("tcp", dest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}
	addr = tcpAddr

	// Retrieve the camera streams
	for _, channel := range channels {
		// TODO: Possible GC'd pointer
		// TODO: Single stream blocks other streams; consider using goroutines and preventing termination
		StreamToStdout(&channel)
	}
}
