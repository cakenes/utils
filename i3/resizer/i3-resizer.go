package main

import (
	"encoding/json"
	"os"
	"os/exec"
)

type Msg struct {
	Success bool `json:"success"`
}

func main() {
	if len(os.Args) < 5 {
		println("Usage: i3-resizer <left|right|up|down> <left|right|up|down> <px> <unit>")
		println("e.g. i3-resizer left right 40 px")
		os.Exit(1)
	}

	args := os.Args[1:5]
	cmd := exec.Command("i3-msg", "resize", "grow", args[0], args[2], args[3])
	out, _ := cmd.Output()

	var msg []Msg

	err := json.Unmarshal(out, &msg)
	if err != nil {
		println("error:", err)
	}

	if !msg[0].Success {
		cmd := exec.Command("i3-msg", "resize", "shrink", args[1], args[2], args[3])
		cmd.Run()
	}
}
