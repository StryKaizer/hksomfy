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
	StepsInMs time.Duration
	Pin       string
	Repeat    int
	Blinds    map[string]blindConfig
}

type blindConfig struct {
	SomfyAddress string
	Label        string
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

	log.Println(config.Pin)
	for blind_name, blind_config := range config.Blinds {
		log.Println(blind_name)
		log.Println(blind_config)
		// TODO: trigger all but last one with go prefix
		initBlind(blind_config)
	}

}

func initBlind(blind_config blindConfig) {
	info := accessory.Info{
		Name:         blind_config.Label,
		SerialNumber: "1.1",
		Model:        "Somfy RTS",
		Manufacturer: "Jimmy Henderickx",
	}

	triggerSomfyCommand("down", blind_config)
	acc = somfaccessory.NewWindowCovering(info)
	acc.WindowCovering.TargetPosition.OnValueRemoteUpdate(func(target int) {

		target_position = target
		current_position := acc.WindowCovering.CurrentPosition.GetValue()
		if current_position < target_position {
			log.Println("New target: " + strconv.Itoa(target_position))
			triggerSomfyCommand("up", blind_config)
		}
		if current_position > target_position {
			log.Println("New target: " + strconv.Itoa(target_position))
			triggerSomfyCommand("down", blind_config)
		}

		if is_moving == false {
			is_moving = true
			go updateCurrentPosition(blind_config)
		}
	})

	log.Println(blind_config.SomfyAddress)
	t, err := hc.NewIPTransport(hc.Config{Pin: config.Pin}, acc.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

	t.Start()

}

func triggerSomfyCommand(command string, blind_config blindConfig) {
	// var command_code int
	var pilight_param string
	switch {
	case command == "up":
		log.Println("Triggering UP for " + blind_config.Label)
		target_direction = "up"
		pilight_param = "-t"
	case command == "down":
		log.Println("Triggering DOWN for " + blind_config.Label)
		target_direction = "down"
		pilight_param = "-f"
	case command == "halt":
		log.Println("Triggering HALT for " + blind_config.Label)
		pilight_param = "-m"
	}

	i := 1
	for i <= config.Repeat {
		cmd := "pilight-send"
		args := []string{"-p", "somfy_rts", "-a", blind_config.SomfyAddress, pilight_param}
		if err := exec.Command(cmd, args...).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Successfully send pilight command")
		i += 1
	}

}

func updateCurrentPosition(blind_config blindConfig) {
	time.Sleep(time.Millisecond * config.StepsInMs)
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
		updateCurrentPosition(blind_config)
	} else {
		is_moving = false
		if new_position > 0 && new_position < 100 {
			triggerSomfyCommand("halt", blind_config)
		}
	}

}
