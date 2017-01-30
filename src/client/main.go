package main

import (
	"os"
	"github.com/namsral/flag"
	"fmt"
	"net"
	"math"
	"strings"
	"strconv"
	"sync"
)

const maxChannels = 4

// Initialize flag variables
var (
	source   *net.TCPAddr
	dest     *net.TCPAddr
	user     string
	pass     string
	channels []int
	wg       sync.WaitGroup
)

func main() {
	// Retrieve the command line flags
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "SWANN", 0)
	fs.StringVar(&user, "user", "", "Username to authenticate with")
	fs.StringVar(&pass, "pass", "", "Password to authenticate with")
	sourceInput := fs.String("source", "", "The address of the DVR in the format host:port")
	destInput := fs.String("dest", "", "The address of the streaming server in the format host:port")
	channelInput := fs.String("channels", "", "Channel(s) to stream, delimited by commas")

	fs.Parse(os.Args[1:])

	// Ensure that the command line flags are not empty
	if user == "" || pass == "" || *sourceInput == "" || *destInput == "" || *channelInput == "" {
		fs.PrintDefaults()
		os.Exit(1)
	}

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
		if intInSlice(&parsedChannel, &channels) {
			fmt.Fprintln(os.Stderr, "All channels need to be unique", maxChannels)
			os.Exit(1)
		}
		channels = append(channels, parsedChannel)
	}

	// Resolve the TCP addresses
	tcpAddr, err := net.ResolveTCPAddr("tcp", *sourceInput)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}
	source = tcpAddr
	tcpAddr, err = net.ResolveTCPAddr("tcp", *destInput)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}
	dest = tcpAddr

	// Retrieve the camera streams
	for _, channel := range channels {
		wg.Add(1)
		go StreamToServer(channel)
	}

	wg.Wait()
}
