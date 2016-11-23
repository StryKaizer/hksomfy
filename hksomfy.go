package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/strykaizer/hksomfy/somfaccessory"
)

type tomlConfig struct {
	step_in_ms     time.Duration
	pin            string
	somfy_address  string
	pilight_repeat int
}


// VARS
var target_position int
var target_direction string
var acc *somfaccessory.WindowCovering
var is_moving bool = false
var config tomlConfig

func main() {

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	info := accessory.Info{
		Name:         "Blinds",
		SerialNumber: "1.1",
		Model:        "Somfy RTS",
		Manufacturer: "Jimmy Henderickx",
	}

	triggerSomfyCommand("down")
	acc = somfaccessory.NewWindowCovering(info)
	acc.WindowCovering.TargetPosition.OnValueRemoteUpdate(func(target int) {

		target_position = target
		current_position := acc.WindowCovering.CurrentPosition.GetValue()
		if current_position < target_position {
			log.Println("New target: " + strconv.Itoa(target_position))
			triggerSomfyCommand("up")
		}
		if current_position > target_position {
			log.Println("New target: " + strconv.Itoa(target_position))
			triggerSomfyCommand("down")
		}

		if is_moving == false {
			is_moving = true
			go updateCurrentPosition()
		}
	})

	t, err := hc.NewIPTransport(hc.Config{Pin: config.pin}, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	t.Start()
}

func triggerSomfyCommand(command string) {
	// var command_code int
	var pilight_param string
	switch {
	case command == "up":
		log.Println("Triggering UP")
		target_direction = "up"
		pilight_param = "-t"
	case command == "down":
		log.Println("Triggering DOWN")
		target_direction = "down"
		pilight_param = "-f"
	case command == "halt":
		log.Println("Triggering HALT")
		pilight_param = "-m"
	}

	i := 1
	for i <= config.pilight_repeat {
		cmd := "pilight-send"
		args := []string{"-p", "somfy_rts", "-a", config.somfy_address, pilight_param}
		if err := exec.Command(cmd, args...).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Successfully send pilight command")
		i += 1
	}

}

func updateCurrentPosition() {
	time.Sleep(time.Millisecond * config.step_in_ms)
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
			triggerSomfyCommand("halt")
		}
	}

}
