package cubase

import (
	"fmt"
	"strconv"
)

type Insert struct {
	Slot       int
	Name       string
	Bypassed   bool
	Parameters []PluginParameter
}

type PluginParameter struct {
	ID              string
	Name            string
	NormalizedValue float64
	DisplayValue    string
}

type InsertCapabilities struct {
	Bypass             bool
	ParameterBankWrite bool
	Discovery          bool
	Readback           bool
}

func (b *Bridge) SetInsertBypass(slot int, bypass bool) error {
	return b.SendInsertBypass(slot, boolValue(bypass))
}

func (b *Bridge) SetInsertParameter(slot int, parameterID string, value float64) error {
	if slot != 1 {
		return fmt.Errorf("parameter bank is currently mapped only for insert slot 1")
	}
	parameter, err := strconv.Atoi(parameterID)
	if err != nil || parameter < 1 || parameter > 8 {
		return fmt.Errorf("parameter ID must be 1..8")
	}
	v, err := normalized(value, 0, 1)
	if err != nil {
		return err
	}
	return b.SendPluginParameter(parameter, v)
}
