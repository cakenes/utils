package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os/exec"
)

type Msg struct {
	Change    string    `json:"change"`
	Container Container `json:"container"`
}

type Container struct {
	WindowRect WindowRect `json:"window_rect"`
}

type WindowRect struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func main() {
	println("Starting i3-autotiler")
	msgEvent := exec.Command("i3-msg", "-t", "subscribe", "-m", "[ \"window\" ]")
	stdout, err := msgEvent.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = msgEvent.Start()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		var msg Msg

		m := scanner.Text()
		err = json.Unmarshal([]byte(m), &msg)
		if err != nil {
			log.Fatal(err)
		}

		if msg.Change == "focus" {
			if msg.Container.WindowRect.Width > msg.Container.WindowRect.Height {
				println("Width", msg.Container.WindowRect.Width, ", Height", msg.Container.WindowRect.Height, " ,Next split will be horizontal")
				exec.Command("i3-msg", "split", "horizontal").Run()
			} else {
				println("Width", msg.Container.WindowRect.Width, ", Height", msg.Container.WindowRect.Height, " ,Next split will be vertical")
				exec.Command("i3-msg", "split", "vertical").Run()
			}
		}
	}

	msgEvent.Wait()
}
