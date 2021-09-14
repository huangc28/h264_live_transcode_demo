# Browser plays h264 stream from golang live transcode server. 

## Why?

## Install & Setup

### Backend


```sh
# Install dependencies.
cd {{PROJ_DIR}} && go get

# Host up project.
make run_local
```

> `{{PROJ_DIR}}` is the path of this project.

2 tcp servers will be up and running on `ws://localhost:3333` and `tcp://localhost:5000`. push your `rawvideo` stream to `:5000`. The tcp server will broadcast the stream received from `:5000` to clients that are connected to `:3333`. 

## Frontend

```
# Download ws-avc-player
cd {{PROJ_DIR}} && git clone git@github.com:matijagaspar/ws-avc-player.git

# Open demo html script
open ./test_browser_play_h264.html 
```

Frontend uses [ws-avc-player](https://github.com/matijagaspar/ws-avc-player) to decode h264 video stream. `ws-avc-player` uses [Broadway.js](https://github.com/mbebenita/Broadway) as the decoder.

## Push stream!

We can now push the stream to transcoding server to test out the browser decoder. We have the following options to test it out.

**ffmpeg**

If your video is in `.mp4` format, you could convert it to [rawvideo](https://stackoverflow.com/questions/7238013/rawvideo-and-rgb32-values-passed-to-ffmpeg#:~:text=%2Df%20rawvideo%20is%20basically%20a,to%20specify%20the%20%2Dpix_fmt%20option.)

```
ffmpeg -r 30 -i ./TESLA-test1-raw.mp4  -vcodec libx264 -vprofile baseline -b:v 500k -bufsize 600k -tune zerolatency -pix_fmt yuv420p -r 15 -g 30 -f rawvideo tcp://localhost:5000
```

**nc**

If you have an existing rawvideo, use `nc` to send stream to `:5000`

```
cat out_5 | nc localhost 5000
```
 