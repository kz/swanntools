package main

import log "github.com/Sirupsen/logrus"

const (
	SaveDiskHandlerType = 1
)

// Consumer is a type which consumes a DVR stream and performs an operation on it
type Consumer struct {
	// Receiver is the channel which receives a byte stream of DVR video
	Receiver chan []byte
	// Handler type is used to figure out which handler to use
	HandlerType int
}

func (c *Consumer) Handle() {
	switch c.HandlerType {
	case SaveDiskHandlerType:
		SaveDisk(&c.Receiver)
	default:
		log.Fatalf("Unknown handler type used: %d\n", c.HandlerType)
	}
}

func SaveDisk(stream *chan []byte) {

}
