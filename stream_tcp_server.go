package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/golang/glog"
)

// TCP server that receives rawVideo stream and broadcast to socket client.
// https://gist.github.com/MilosSimic/ae7fe8d70866e89dbd6e84d86dc8d8d5
const (
	HOST = "localhost"
	PORT = "5000"
	TYPE = "tcp"
)

func RunStreamTcpServer(hub *Hub) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", HOST, PORT))

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		c, err := l.Accept()

		if err != nil {
			return err
		}

		go handleConnection(c, hub)
	}
}

func broadcastToClient(stream io.ReadCloser, hub *Hub) error {

	// Make a 32mb buffer.
	buf := make([]byte, 32*1024)

	for {
		rnum, err := stream.Read(buf)

		if err != nil {
			return err
		}

		if rnum > 0 {
			// Stream bytes to all socket client, so we need the hub to publish stream to client connections.
			newb := make([]byte, rnum)
			copy(newb, buf[0:rnum])

			hub.broadcast <- newb
		}

	}
}

func handleConnection(c net.Conn, hub *Hub) {
	log.Printf("Serving %s\n", c.RemoteAddr().String())

	for {
		evtReader := NewEventReader(c)

		stream := evtReader.ReadEvent()

		err := broadcastToClient(stream, hub)

		// Make a 1mb buffer.
		if err == io.EOF {
			glog.Infof("end of file %v", err)

			break
		}

		if err != nil {
			glog.Infof("failed to broadcast to client %v", err)

			break
		}
	}

	c.Close()
}
