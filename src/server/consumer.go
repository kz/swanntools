package main

import (
	log "github.com/Sirupsen/logrus"
	"time"
	"os"
	"strconv"
)

const (
	SaveDiskHandlerType = 1
)

// Data is a struct which contains the channel number and stream of the data being sent
type Data struct {
	channel int
	stream  []byte
}

// Consumer is a type which consumes a DVR stream and performs an operation on it
type Consumer struct {
	// Receiver is the channel which receives a byte stream of DVR video
	Receiver chan Data
	// Handler is used to figure out which handler to use
	HandlerType int
	// Destination is the destination (e.g., file path to directory) of the stream
	Destination string
}

func (c *Consumer) Handle() {
	switch c.HandlerType {
	// Sends data to be saved on disk
	case SaveDiskHandlerType:
		for {
			select {
			case data := <-c.Receiver:
				c.saveDisk(data)
			}
		}
	default:
		log.Fatalf("Unknown handler type used: %d\n", c.HandlerType)
	}
}

// saveDisk saves the stream to a file which rotates every hour
func (c *Consumer) saveDisk(data Data) {
	// Generate file name
	fileName := time.Now().Format("2006-01-02-15-") + strconv.Itoa(data.channel) + ".h264"

	// Generate file path
	path := c.Destination + "/" + fileName

	// Open file path
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.WithField("Path", path).Fatalln("Unable to open file: ", err.Error())
	}
	defer f.Close()

	// Write to file
	_, err = f.Write(data.stream)
	if err != nil {
		log.WithField("Path", path).Fatalln("Error when writing to file: ", err.Error())
	}
}
