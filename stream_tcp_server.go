package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"

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
	out, err := os.OpenFile("./out_4", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	defer out.Close()

	// Make a 32mb buffer.
	buf := make([]byte, 32*1024)

	for {
		rnum, err := stream.Read(buf)

		// if err == io.EOF {
		// 	close(hub.broadcast)
		// }

		if err != nil {
			return err
		}

		if rnum > 0 {
			// glog.Infof("DEBUG byte length %v", rn)
			// glog.Infof("DEBUG stream tcp socket client %v", hub.clients)

			// Stream bytes to all socket client, so we need the hub to publish stream to client connections.
			// Iterate through hub clients
			// log.Printf("DEBUG 2 trigger send!!! %v", hub.clients)

			br := bytes.NewReader(buf[0:rnum])

			if _, err := io.Copy(out, br); err != nil {
				return err
			}

			newb := make([]byte, rnum)
			copy(newb, buf[0:rnum])

			// hub.broadcast <- buf[0:rnum]
			hub.broadcast <- newb
		}

	}
}

func handleConnection(c net.Conn, hub *Hub) {
	log.Printf("Serving %s\n", c.RemoteAddr().String())

	for {
		// netData, err := bufio.NewReader(c).ReadString('\n')
		// bytes := bufio.NewScanner(c).Scan()

		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// log.Printf("DEBUG 1 net data length %v", len(netData))
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

		// log.Printf("DEBUG 2 net data length %v", len(bytes))

		// temp := strings.TrimSpace(string(bytes))
		// if temp == "STOP" {
		// 	break
		// }

		// We will broadcast the byte data to socket client.
		// c.Write([]byte(string([]byte{123})))
	}

	c.Close()
}
