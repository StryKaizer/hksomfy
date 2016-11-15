package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/strykaizer/hksomfy/somfaccessory"
)

// CONFIG
var step_in_ms time.Duration = 184 // miliseconds per step, x 100 is entire run
var pin int = 11111111

// VARS
var target_position int
var target_direction string
var acc *somfaccessory.WindowCovering
var is_moving bool = false

func main() {

	info := accessory.Info{
		Name:         "Blinds", // Do not use spaces here, get transformed to _.
		SerialNumber: "1.0",
		Model:        "Somfy RTS",
		Manufacturer: "Jimmy Henderickx",
	}

	triggerSomfyCommand("down")
	acc = somfaccessory.NewWindowCovering(info)
	acc.WindowCovering.TargetPosition.OnValueRemoteUpdate(func(target int) {

		target_position = target
		current_position := acc.WindowCovering.CurrentPosition.GetValue()
		if current_position < target_position {
			log.Println("Triggering UP. Target: " + strconv.Itoa(target_position))
			// TODO: Trigger up here
			triggerSomfyCommand("up")
			triggerSomfyCommand("up")
			triggerSomfyCommand("up")
			triggerSomfyCommand("up")
			target_direction = "up"
		}
		if current_position > target_position {
			log.Println("Triggering DOWN. Target: " + strconv.Itoa(target_position))
			// TODO: Trigger down here
			triggerSomfyCommand("down")
			triggerSomfyCommand("down")
			triggerSomfyCommand("down")
			triggerSomfyCommand("down")
			target_direction = "down"
		}

		if is_moving == false {
			is_moving = true
			go updateCurrentPosition()
		}
	})

	t, err := hc.NewIPTransport(hc.Config{Pin: "11111111"}, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	t.Start()
}

func triggerSomfyCommand(command string) {
	var command_code int
	switch {
	case command == "up":
		command_code = 2
	case command == "down":
		command_code = 4
	case command == "halt":
		command_code = 1
	}
	conn, err := net.Dial("tcp", "192.168.1.180:5000")
	reply := make([]byte, 1024)
	strEcho := "{\"action\": \"send\", \"code\": {\"protocol\": [\"somfy_rts\"],	\"address\": 2235423, \"command_code\": " + strconv.Itoa(command_code) + "}}"
	_, err = conn.Write([]byte(strEcho))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	println("write to server = ", strEcho)

	reply = make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}
}

func updateCurrentPosition() {
	time.Sleep(time.Millisecond * step_in_ms)
	current_position := acc.WindowCovering.CurrentPosition.GetValue()
	var new_position int
	if target_direction == "up" {
		new_position = current_position + 1
	} else {
		new_position = current_position - 1
	}
	acc.WindowCovering.CurrentPosition.SetValue(new_position)
	log.Println("Current position: " + strconv.Itoa(new_position))

	if new_position != target_position && new_position > 0 && new_position < 100 {
		updateCurrentPosition()
	} else {
		is_moving = false
		if new_position > 0 && new_position < 100 {
			// TODO: Trigger halt here.
			triggerSomfyCommand("halt")
			triggerSomfyCommand("halt")
			triggerSomfyCommand("halt")
			triggerSomfyCommand("halt")
			log.Println("Triggering HALT")
		}
	}

}
