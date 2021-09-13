package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

// We will start a websocket that receives video stream from ffmpeg command:
//
//   ffmpeg -r 30 -i ./TESLA-test1-raw.mp4  -vcodec libx264 -vprofile baseline -b:v 500k -bufsize 600k -tune zerolatency -pix_fmt yuv420p -r 15 -g 30 -f rawvideo tcp://localhost:5000
//
// We will broadcast the video stream buffer to websocket client, in this case, the web. Web uses `https://github.com/matijagaspar/ws-avc-player ` to decode
// [rawvideo](https://stackoverflow.com/questions/7238013/rawvideo-and-rgb32-values-passed-to-ffmpeg) format to browser compatible h264 video stream.
// `ws-avc-player` uses [broadway.js](https://github.com/mbebenita/Broadway) underneath it's implementation.

func main() {
	flag.Parse()
	// Initialize hub in a gorouting to handle following jobs in the background:
	//   - client (connection) register.
	//   - client (connection) unregister.
	//   - broadcast message to everyone(other clients) in the chatroom.
	hub := newHub()
	go hub.run()

	videoStreamWSServerHandler := mux.NewRouter()
	videoStreamWSServerHandler.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		broadcastVideoStream(hub, w, r)
	})

	videoStreamWSServer := &http.Server{
		Handler:      videoStreamWSServerHandler,
		Addr:         ":3333",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Create a socket server listen to port :3333 to broadcast video bytes to client
	go func() {
		glog.Info("video stream server listen on :3333")

		if err := videoStreamWSServer.ListenAndServe(); err != nil {
			glog.Fatal(err)
		}
	}()

	go func(hub *Hub) {
		glog.Info("video stream receiver listen on :5000")

		if err := RunStreamTcpServer(hub); err != nil {
			glog.Fatal(err)
		}
	}(hub)

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c
	ctx := context.Background()

	videoStreamWSServer.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
