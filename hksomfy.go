package main

import (
	"log"
	"strconv"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/strykaizer/hksomfy/somfaccessory"
)

func main() {
	info := accessory.Info{
		Name:         "Blinds", // Do not use spaces here, get transformed to _.
		SerialNumber: "1.0",
		Model:        "Somfy RTS",
		Manufacturer: "Jimmy Henderickx",
	}

	acc := somfaccessory.NewWindowCovering(info)

	acc.WindowCovering.CurrentPosition.OnValueRemoteUpdate(func(currentposition int) {
		log.Println("CurrentPosition" + strconv.Itoa(currentposition))
	})

	acc.WindowCovering.TargetPosition.OnValueRemoteUpdate(func(targetposition int) {
		log.Println("TargetPosition" + strconv.Itoa(targetposition))
		acc.WindowCovering.CurrentPosition.SetValue(100) // TODO
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
