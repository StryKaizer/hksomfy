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
var step_in_ms time.Duration = 184        // Miliseconds per step, x 100 is entire run (From closed to open)
var pin string = "11111111"               // Pincode used in Homekit
var pilight_host string = "192.168.1.180" // Hostname pilight daemon
var pilight_port string = "5000"          // Port pilight daemon
var pilight_repeat int = 4                // Times a command is executed, for less stable connections support.

// VARS
var target_position int
var target_direction string
var acc *somfaccessory.WindowCovering
var is_moving bool = false

func main() {

	info := accessory.Info{
		Name:         "Blinds",
		SerialNumber: "1.0",
		Model:        "Somfy RTS",
		Manufacturer: "Jimmy Henderickx",
	}

	go triggerSomfyCommand("down")
	acc = somfaccessory.NewWindowCovering(info)
	acc.WindowCovering.TargetPosition.OnValueRemoteUpdate(func(target int) {

		target_position = target
		current_position := acc.WindowCovering.CurrentPosition.GetValue()
		if current_position < target_position {
			log.Println("New target: " + strconv.Itoa(target_position))
			go triggerSomfyCommand("up")
		}
		if current_position > target_position {
			log.Println("New target: " + strconv.Itoa(target_position))
			go triggerSomfyCommand("down")
		}

		if is_moving == false {
			is_moving = true
			go updateCurrentPosition()
		}
	})

	t, err := hc.NewIPTransport(hc.Config{Pin: pin}, acc.Accessory)
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
		log.Println("Triggering UP")
		target_direction = "up"
		command_code = 2
	case command == "down":
		log.Println("Triggering DOWN")
		target_direction = "down"
		command_code = 4
	case command == "halt":
		log.Println("Triggering HALT")
		command_code = 1
	}

	i := 1
	for i <= pilight_repeat {
		conn, err := net.Dial("tcp", pilight_host+":"+pilight_port)
		reply := make([]byte, 1024)
		strEcho := "{\"action\": \"send\", \"code\": {\"protocol\": [\"somfy_rts\"],	\"address\": 2235423, \"command_code\": " + strconv.Itoa(command_code) + "}}"
		_, err = conn.Write([]byte(strEcho))
		if err != nil {
			println("Write to server failed:", err.Error())
			os.Exit(1)
		}

		reply = make([]byte, 1024)

		_, err = conn.Read(reply)
		if err != nil {
			println("Write to server failed:", err.Error())
			os.Exit(1)
		}
		i += 1
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
			go triggerSomfyCommand("halt")
		}
	}

}
