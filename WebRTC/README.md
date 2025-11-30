# WebRTC Video Streaming Server

This project is a simple WebRTC server written in Go that streams a video file (`video.ivf`) to connected clients. It uses the `pion/webrtc` library.

## Prerequisites

-   [Go](https://go.dev/dl/) (version 1.20 or later recommended)
-   [FFmpeg](https://ffmpeg.org/download.html) (for generating the compatible video file)

## Installation

1.  Clone the repository (if you haven't already).
2.  Navigate to the project directory:
    ```bash
    cd /path/to/WebRTC
    ```
3.  Install the Go dependencies:
    ```bash
    go mod download
    ```

## Video and Audio Setup

The server requires a VP8 encoded IVF file (`video.ivf`) and an Opus encoded Ogg file (`audio.ogg`) in the root directory.

### Generate Video (`video.ivf`)

To generate a high-quality video file from an existing video (e.g., `input.mp4`), use the following `ffmpeg` command:

```bash
ffmpeg -i input.mp4 -g 30 -vcodec libvpx -b:v 2500k -crf 10 -f ivf video.ivf
```

-   `-g 30`: Sets the keyframe interval.
-   `-vcodec libvpx`: Uses the VP8 video codec.
-   `-b:v 2500k`: Sets the target video bitrate to 2.5 Mbps (adjust as needed).
-   `-crf 10`: Sets the Constant Rate Factor for quality (lower is better).
-   `-f ivf`: Sets the output format to IVF.

### Generate Audio (`audio.ogg`)

To generate the audio file:

```bash
ffmpeg -i input.mp4 -vn -c:a libopus -page_duration 20000 -f ogg audio.ogg
```

-   `-vn`: No video.
-   `-c:a libopus`: Uses the Opus audio codec.
-   `-page_duration 20000`: Sets the Ogg page duration to 20ms (important for WebRTC).
-   `-f ogg`: Sets the output format to Ogg.

**Note:** Ensure the output files are named exactly `video.ivf` and `audio.ogg` and are placed in the same directory as `main.go`.

## Running the Server

1.  Start the server:
    ```bash
    go run main.go
    ```
2.  The server will start on `http://localhost:8080`.
3.  Open a WebRTC compatible browser and navigate to `http://localhost:8080`.
4.  The video stream should start playing automatically once the connection is established.

## Project Structure

-   `main.go`: The main server application code. Handles HTTP signaling and WebRTC streaming.
-   `go.mod` / `go.sum`: Go module definitions and dependencies.
-   `static/`: Directory for static assets (HTML/JS client).
-   `video.ivf`: The source video file (must be generated).

## Troubleshooting

-   **"open video.ivf: no such file or directory"**: You missed the Video Setup step. Please generate or place a `video.ivf` file in the project root.
-   **Connection Failed**: Check your browser console for errors. Ensure no firewalls are blocking the connection.
