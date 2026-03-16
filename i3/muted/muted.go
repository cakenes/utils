package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func writeState(state string) {
	os.WriteFile("/dev/shm/mic_state", []byte(state+"\n"), 0644)
}

func main() {
	args := os.Args[1:2]

	lowThreshold := 0.00003
	highThreshold := 0.00010
	state := "Unknown"

	if len(args) > 0 {
		if value, err := strconv.ParseFloat(args[0], 64); err == nil {
			lowThreshold = value
		}
	}
	if len(args) > 1 {
		if value, err := strconv.ParseFloat(args[1], 64); err == nil {
			highThreshold = value
		}
	}

	rmsRegex := regexp.MustCompile(`RMS\s+amplitude:\s+([0-9.]+)`)

	for {
		// Record 0.2s into RAM
		exec.Command("timeout", "0.2", "arecord", "-f", "cd", "-q", "/dev/shm/mic_test.wav").Run()

		// Run sox stat
		cmd := exec.Command("sox", "/dev/shm/mic_test.wav", "-n", "stat")
		stderr := &bytes.Buffer{}
		cmd.Stdout = nil
		cmd.Stderr = stderr
		cmd.Run()

		// Extract amplitude
		scanner := bufio.NewScanner(strings.NewReader(stderr.String()))
		rms := 0.0
		for scanner.Scan() {
			line := scanner.Text()
			if rmsRegex.MatchString(line) {
				fmt.Sscanf(rmsRegex.FindStringSubmatch(line)[1], "%f", &rms)
				break
			}
		}

		newState := state
		if rms < lowThreshold {
			newState = "Muted"
		} else if rms > highThreshold {
			newState = "Active"
		}

		// Write only on change
		if newState != state {
			writeState(newState)
			state = newState
		}

		time.Sleep(100 * time.Millisecond)
	}
}
