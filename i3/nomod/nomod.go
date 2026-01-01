package main

import (
    "log"
    "regexp"
    "strings"
    i3 "go.i3wm.org/i3/v4"
)

var gameList = []string{
    "cs2",
    "steam_app_.*",
    "Wine",
}

var gameRegex = regexp.MustCompile("(?i)(" + strings.Join(gameList, "|") + ")")

func main() {
    // Subscribe returns *EventReceiver
    ch := i3.Subscribe(i3.WindowEventType)

    for ch.Next() {
        // Receive next event
        event := ch.Event()

        win, ok := event.(*i3.WindowEvent)
        if !ok || win.Change != "focus" {
            continue
        }

        tree, err := i3.GetTree()
        if err != nil {
            log.Println("Error getting tree:", err)
            continue
        }

        focused := tree.Root.FindFocused(func(n *i3.Node) bool {
            return n.Focused
        })

        if focused == nil {
            continue
        }

        props := focused.WindowProperties
        title := props.Title
        class := props.Class

        var cmd string
        if gameRegex.MatchString(title) || gameRegex.MatchString(class) {
            log.Println("Entering nomod mode for:", title, class)
            cmd = `mode "Nomod"`
        } else {
            log.Println("Returning to default mode for:", title, class)
            cmd = `mode "default"`
        }

        results, err := i3.RunCommand(cmd)
        if err != nil {
            log.Println("Error running command:", err)
        }
        for _, res := range results {
            if !res.Success {
                log.Println("i3 command error:", res.Error)
            }
        }
    }
}
