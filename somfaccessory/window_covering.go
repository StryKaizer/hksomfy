package somfaccessory

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type WindowCovering struct {
	Accessory      *accessory.Accessory
	WindowCovering *service.WindowCovering
}

// NewWindowCovering returns a window covering accessory which one window covering service.
func NewWindowCovering(info accessory.Info) *WindowCovering {
	acc := WindowCovering{}
	acc.Accessory = accessory.New(info, accessory.TypeWindowCovering)
	acc.WindowCovering = service.NewWindowCovering()

	acc.WindowCovering.CurrentPosition.SetValue(0)

	acc.Accessory.AddService(acc.WindowCovering.Service)

	return &acc
}
