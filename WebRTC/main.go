package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/ivfreader"
	"github.com/pion/webrtc/v4/pkg/media/oggreader"
)

var (
	peerConnections   = make(map[string]*webrtc.PeerConnection)
	candidatesMux     sync.RWMutex
	pendingCandidates = make(map[string][]webrtc.ICECandidateInit)
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/offer", handleOffer)
	http.HandleFunc("/candidate", handleCandidate)
	http.HandleFunc("/answer-candidates", handleAnswerCandidates)

	fmt.Println("WebRTC server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleOffer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type OfferRequest struct {
		SDP       webrtc.SessionDescription `json:"sdp"`
		SessionID string                    `json:"sessionId"`
	}

	var offerReq OfferRequest
	if err := json.Unmarshal(body, &offerReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create peer connection with STUN and TURN servers
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
			{
				URLs: []string{"stun:stun1.l.google.com:19302"},
			},
			// Free TURN server (for testing only - use your own in production)
			{
				URLs:       []string{"turn:openrelay.metered.ca:80"},
				Username:   "openrelayproject",
				Credential: "openrelayproject",
			},
		},
	}

	// Create settings for the peer connection
	settingEngine := webrtc.SettingEngine{}

	// Use mDNS to help with local network connections
	settingEngine.SetICEMulticastDNSMode(ice.MulticastDNSModeQueryOnly)

	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))
	pc, err := api.NewPeerConnection(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store the peer connection
	candidatesMux.Lock()
	peerConnections[offerReq.SessionID] = pc
	pendingCandidates[offerReq.SessionID] = make([]webrtc.ICECandidateInit, 0)
	candidatesMux.Unlock()

	// Handle ICE connection state
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("[%s] ICE Connection State: %s\n", offerReq.SessionID, state.String())

		if state == webrtc.ICEConnectionStateFailed ||
			state == webrtc.ICEConnectionStateDisconnected {
			candidatesMux.Lock()
			delete(peerConnections, offerReq.SessionID)
			delete(pendingCandidates, offerReq.SessionID)
			candidatesMux.Unlock()
		}
	})

	// Store ICE candidates as they're gathered
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateInit := candidate.ToJSON()
		log.Printf("[%s] Server ICE Candidate: %s\n", offerReq.SessionID, candidateInit.Candidate)

		candidatesMux.Lock()
		pendingCandidates[offerReq.SessionID] = append(
			pendingCandidates[offerReq.SessionID],
			candidateInit,
		)
		candidatesMux.Unlock()
	})

	// Handle incoming tracks
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("[%s] Track received: %s\n", offerReq.SessionID, track.Kind())

		go func() {
			for {
				_, _, readErr := track.ReadRTP()
				if readErr != nil {
					return
				}
			}
		}()
	})

	// Create video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		"video",
		"pion-video",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rtpSender, err := pc.AddTrack(videoTrack)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Create audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"pion-audio",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	audioSender, err := pc.AddTrack(audioTrack)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := audioSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Set remote description
	if err := pc.SetRemoteDescription(offerReq.SDP); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set local description
	if err := pc.SetLocalDescription(answer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start sending video
	go sendVideoFrames(videoTrack, offerReq.SessionID)

	// Start sending audio
	go sendAudioFrames(audioTrack, offerReq.SessionID)

	// Return answer
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sdp":       answer,
		"sessionId": offerReq.SessionID,
	})
}

func handleCandidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	type CandidateRequest struct {
		Candidate webrtc.ICECandidateInit `json:"candidate"`
		SessionID string                  `json:"sessionId"`
	}

	var candReq CandidateRequest
	if err := json.Unmarshal(body, &candReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	candidatesMux.RLock()
	pc, exists := peerConnections[candReq.SessionID]
	candidatesMux.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if err := pc.AddICECandidate(candReq.Candidate); err != nil {
		log.Printf("[%s] Error adding ICE candidate: %v\n", candReq.SessionID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[%s] Added client ICE candidate: %s\n", candReq.SessionID, candReq.Candidate.Candidate)
	w.WriteHeader(http.StatusOK)
}

func handleAnswerCandidates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId required", http.StatusBadRequest)
		return
	}

	candidatesMux.Lock()
	candidates := pendingCandidates[sessionID]
	pendingCandidates[sessionID] = make([]webrtc.ICECandidateInit, 0)
	candidatesMux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"candidates": candidates,
	})
}

func sendVideoFrames(track *webrtc.TrackLocalStaticSample, sessionID string) {
	// Open a IVF file and start reading using our IVFReader
	file, err := os.Open("assets/video.ivf")
	if err != nil {
		log.Printf("[%s] Error opening video.ivf: %v\n", sessionID, err)
		return
	}

	ivf, header, err := ivfreader.NewWith(file)
	if err != nil {
		log.Printf("[%s] Error creating IVF reader: %v\n", sessionID, err)
		file.Close()
		return
	}

	// Determine the ticker interval from the header
	// TimebaseNumerator / TimebaseDenominator gives the duration in seconds per frame
	// We want duration in milliseconds
	tickerDuration := time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000)
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	log.Printf("[%s] Started sending video frames from video.ivf\n", sessionID)

	for range ticker.C {
		candidatesMux.RLock()
		_, exists := peerConnections[sessionID]
		candidatesMux.RUnlock()

		if !exists {
			log.Printf("[%s] Session closed, stopping video\n", sessionID)
			file.Close()
			return
		}

		frame, _, err := ivf.ParseNextFrame()
		if err == io.EOF {
			log.Printf("[%s] End of video file, restarting...\n", sessionID)
			file.Close()
			file, err = os.Open("assets/video.ivf")
			if err != nil {
				log.Printf("[%s] Error re-opening video.ivf: %v\n", sessionID, err)
				return
			}
			ivf, _, err = ivfreader.NewWith(file)
			if err != nil {
				log.Printf("[%s] Error re-creating IVF reader: %v\n", sessionID, err)
				file.Close()
				return
			}
			continue
		}

		if err != nil {
			log.Printf("[%s] Error parsing frame: %v\n", sessionID, err)
			file.Close()
			return
		}

		if err := track.WriteSample(media.Sample{
			Data:     frame,
			Duration: tickerDuration,
		}); err != nil {
			log.Printf("[%s] Error writing sample: %v\n", sessionID, err)
			file.Close()
			return
		}
	}
}
func sendAudioFrames(track *webrtc.TrackLocalStaticSample, sessionID string) {
	// Open an Ogg file and start reading using our OggReader
	file, err := os.Open("assets/audio.ogg")
	if err != nil {
		log.Printf("[%s] Error opening audio.ogg: %v\n", sessionID, err)
		return
	}

	ogg, _, err := oggreader.NewWith(file)
	if err != nil {
		log.Printf("[%s] Error creating Ogg reader: %v\n", sessionID, err)
		file.Close()
		return
	}

	// Keep track of last granule, the difference is the amount of samples in the buffer
	var lastGranule uint64

	// It is important to use a ticker to pace the audio frames
	// Ogg pages typically contain 20ms of audio
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	log.Printf("[%s] Started sending audio frames from audio.ogg\n", sessionID)

	for range ticker.C {
		candidatesMux.RLock()
		_, exists := peerConnections[sessionID]
		candidatesMux.RUnlock()

		if !exists {
			log.Printf("[%s] Session closed, stopping audio\n", sessionID)
			file.Close()
			return
		}

		pageData, pageHeader, err := ogg.ParseNextPage()
		if err == io.EOF {
			log.Printf("[%s] End of audio file, restarting...\n", sessionID)
			file.Close()
			file, err = os.Open("assets/audio.ogg")
			if err != nil {
				log.Printf("[%s] Error re-opening audio.ogg: %v\n", sessionID, err)
				return
			}
			ogg, _, err = oggreader.NewWith(file)
			if err != nil {
				log.Printf("[%s] Error re-creating Ogg reader: %v\n", sessionID, err)
				file.Close()
				return
			}
			lastGranule = 0
			continue
		}

		if err != nil {
			log.Printf("[%s] Error parsing audio page: %v\n", sessionID, err)
			file.Close()
			return
		}

		// The amount of samples is the difference between the last and current timestamp
		sampleCount := float64(pageHeader.GranulePosition - lastGranule)
		lastGranule = pageHeader.GranulePosition
		sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

		if err := track.WriteSample(media.Sample{
			Data:     pageData,
			Duration: sampleDuration,
		}); err != nil {
			log.Printf("[%s] Error writing audio sample: %v\n", sessionID, err)
			file.Close()
			return
		}
	}
}
