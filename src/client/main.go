package main

import (
	"os"
	"github.com/namsral/flag"
	"fmt"
	"net"
	"math"
)

// Initialize flag variables
var (
	addr *net.TCPAddr
	dest string
	user string
	pass string
	// TODO: global channel variable will be deprecated in favor of allowing multiple channel streams
	channel int
)

func main() {
	// Retrieve the command line flags
	fs := flag.NewFlagSetWithEnvPrefix(os.Args[0], "SWANN", 0)
	fs.StringVar(&dest, "dest", "", "The destination of the DVR in the format host:port")
	fs.StringVar(&user, "user", "", "Username to authenticate with")
	fs.StringVar(&pass, "pass", "", "Password to authenticate with")
	fs.IntVar(&channel, "channel", 0, "Channel to stream, either 1, 2, 3 or 4")
	fs.Parse(os.Args[1:])

	// Ensure that the command line flags are valid
	if dest == "" || user == "" || pass == "" || !(channel == 1 || channel == 2 || channel == 3 || channel == 4) {
		fs.PrintDefaults()
		os.Exit(1)
	}
	// Convert channel from 1, 2, 3, 4 to 1, 2, 4, 8 respectively
	channel = int(math.Exp2(float64(channel - 1)))

	// Resolve the TCP address
	fmt.Fprintln(os.Stdout, "Resolving TCP address.")
	tcpAddr, err := net.ResolveTCPAddr("tcp", dest)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ResolveTCPAddr failed: ", err.Error())
		os.Exit(1)
	}
	addr = tcpAddr

	// TODO: Handle multiple channels
	// Retrieve the camera stream
	StreamToStdout(&channel)
}
