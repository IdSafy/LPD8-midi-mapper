package main

import "math"

// CCTransformConfig describes a value transformation applied to a specific CC number.
// MIDI CC values are always 0-127 (7-bit). InputMin/InputMax and OutputMin/OutputMax
// default to 0 and 127 respectively if left as zero values.
//
// Supported types:
//   - "log"    : linear input → logarithmic output (mimics audio taper potentiometers)
//   - "exp"    : linear input → exponential output (inverse of log)
//   - "remap"  : linear remap from input range to output range
//   - "invert" : output = OutputMax - (input - InputMin); flips the range
type CCTransformConfig struct {
	CC        uint8   `json:"cc"`
	Type      string  `json:"type"`
	InputMin  float64 `json:"input_min"`
	InputMax  float64 `json:"input_max"`
	OutputMin float64 `json:"output_min"`
	OutputMax float64 `json:"output_max"`
	// Base is used by "log" and "exp" transforms. Higher values = more curve. Default: 10.
	Base float64 `json:"base"`
}

func (c *CCTransformConfig) effectiveBase() float64 {
	if c.Base <= 1 {
		return 10
	}
	return c.Base
}

func (c *CCTransformConfig) effectiveInputMin() float64  { return c.InputMin }
func (c *CCTransformConfig) effectiveInputMax() float64 {
	if c.InputMax == 0 {
		return 127
	}
	return c.InputMax
}
func (c *CCTransformConfig) effectiveOutputMin() float64  { return c.OutputMin }
func (c *CCTransformConfig) effectiveOutputMax() float64 {
	if c.OutputMax == 0 {
		return 127
	}
	return c.OutputMax
}

// applyTransform applies the configured transform to a raw MIDI CC value (0-127)
// and returns the transformed value clamped to 0-127.
func applyTransform(value uint8, cfg CCTransformConfig) uint8 {
	inMin := cfg.effectiveInputMin()
	inMax := cfg.effectiveInputMax()
	outMin := cfg.effectiveOutputMin()
	outMax := cfg.effectiveOutputMax()

	v := float64(value)

	// clamp input to declared range
	v = math.Max(inMin, math.Min(inMax, v))

	// normalize to [0, 1]
	var t float64
	if inMax == inMin {
		t = 0
	} else {
		t = (v - inMin) / (inMax - inMin)
	}

	// apply curve
	switch cfg.Type {
	case "log":
		// t_out = log(1 + t*(base-1)) / log(base)
		base := cfg.effectiveBase()
		t = math.Log(1+t*(base-1)) / math.Log(base)
	case "exp":
		// inverse of log: t_out = (base^t - 1) / (base - 1)
		base := cfg.effectiveBase()
		t = (math.Pow(base, t) - 1) / (base - 1)
	case "invert":
		t = 1 - t
	case "remap":
		// pure linear remap; t stays as-is
	default:
		// unknown type: pass through unchanged
		return value
	}

	// scale to output range
	out := outMin + t*(outMax-outMin)

	// clamp to valid MIDI range
	out = math.Max(0, math.Min(127, out))
	return uint8(math.Round(out))
}

// buildCCTransformMap indexes the transform list by CC number for O(1) lookup.
func buildCCTransformMap(transforms []CCTransformConfig) map[uint8]CCTransformConfig {
	m := make(map[uint8]CCTransformConfig, len(transforms))
	for _, t := range transforms {
		m[t.CC] = t
	}
	return m
}
