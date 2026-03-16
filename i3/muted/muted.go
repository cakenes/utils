package main

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"math"
	"os"
	"time"
)

func writeState(state string) {
	os.WriteFile("/dev/shm/mic_state", []byte(state+"\n"), 0644)
}

func getRmsAmplitude(samples []int16) float64 {
	var sum float64
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}
	return math.Sqrt(sum / float64(len(samples)))
}

func isVirtualDevice(name string) bool {
	virtual := []string{
		"default",
		"pulse",
		"pipewire",
		"dmix",
		"sysdefault",
		"front",
		"surround",
		"iec958",
		"spdif",
		"samplerate",
		"speex",
		"lavrate",
		"upmix",
		"vdownmix",
	}
	for _, v := range virtual {
		if name == v {
			return true
		}
	}
	return false
}

func main() {
	lowThreshold := 0.8
	highThreshold := 1.0
	state := "Unknown"

	portaudio.Initialize()
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		panic(err)
	}

	var mic *portaudio.DeviceInfo

	for _, d := range devices {
		if d.MaxInputChannels > 0 && !isVirtualDevice(d.Name) {
			mic = d
			break
		}
	}

	if mic == nil {
		panic("No real input device found")
	}

	fmt.Println("Using device:", mic.Name)

	// Large buffer to avoid overflow
	in := make([]int16, 8192)

	params := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   mic,
			Channels: 1,
			Latency:  mic.DefaultLowInputLatency,
		},
		SampleRate:      48000,
		FramesPerBuffer: len(in),
	}

	stream, err := portaudio.OpenStream(params, in)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	stream.Start()
	defer stream.Stop()

	for {
		if err := stream.Read(); err != nil {
			fmt.Println("Audio read error:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		rms := getRmsAmplitude(in)

		newState := state
		if rms < lowThreshold {
			newState = "Muted"
		} else if rms > highThreshold {
			newState = "Active"
		}

		if newState != state {
			writeState(newState)
			state = newState
		}

		time.Sleep(100 * time.Millisecond)
	}
}
