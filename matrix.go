package main

// RGB represents an RGB color.
type RGB struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

func OneByteToTwoMidiBytes(b uint8) (uint8, uint8) {
	return b >> 4, b & 0x0F
}

func (rgb *RGB) ToAkaiBytes() []byte {
	redHigh, redLow := OneByteToTwoMidiBytes(rgb.Red)
	greenHigh, greenLow := OneByteToTwoMidiBytes(rgb.Green)
	blueHigh, blueLow := OneByteToTwoMidiBytes(rgb.Blue)
	return []byte{redHigh, redLow, greenHigh, greenLow, blueHigh, blueLow}
}

// PadState represents a single element in the matrix.
type PadState struct {
	On               bool // State of the element (on or off)
	Color            RGB  // RGB color of the element
	AlternativeColor RGB  // RGB color used when pressed
	Reverse          bool
	Static           bool
}

// PadsStateArray represents an n x m matrix.
type PadsStateArray struct {
	Elements []PadState // 2D slice to store matrix elements
}

// NewPadsStateArray creates a new n x m matrix with all elements turned off and set to default color.
func NewPadsStateArray(lenght int, pads_config []PadConfig) *PadsStateArray {
	array := &PadsStateArray{
		Elements: make([]PadState, lenght),
	}
	for i := 0; i < lenght; i++ {
		array.Elements[i] = PadState{
			On:               false,
			Color:            RGB{pads_config[i].RGB[0], pads_config[i].RGB[1], pads_config[i].RGB[2]},
			AlternativeColor: RGB{pads_config[i].AlternativeRGB[0], pads_config[i].AlternativeRGB[1], pads_config[i].AlternativeRGB[2]},
			Reverse:          pads_config[i].Reverse,
			Static:           pads_config[i].Static,
		}
	}
	return array
}

// TurnOn sets the element at (row, col) to "on".
func (m *PadsStateArray) TurnOn(index int) {
	if m.isValidPosition(index) {
		m.Elements[index].On = true
	}
}

// TurnOff sets the element at (row, col) to "off".
func (m *PadsStateArray) TurnOff(index int) {
	if m.isValidPosition(index) {
		m.Elements[index].On = false
	}
}

// TurnOff sets the element at (row, col) to "off".
func (m *PadsStateArray) Switch(index int) {
	if m.isValidPosition(index) {
		m.Elements[index].On = !m.Elements[index].On
	}
}

// SetColor sets the RGB color of the element at (row, col).
func (m *PadsStateArray) SetColor(index int, color RGB) {
	if m.isValidPosition(index) {
		m.Elements[index].Color = color
	}
}

// GetElement returns the MatrixElement at (row, col).
func (m *PadsStateArray) GetElement(index int) *PadState {
	if m.isValidPosition(index) {
		return &m.Elements[index]
	}
	return nil
}

// isValidPosition checks if the given row and column are within the matrix bounds.
func (m *PadsStateArray) isValidPosition(index int) bool {
	return index >= 0 && index < len(m.Elements)
}

func (m *PadsStateArray) ToColorChangeSysEx(kind string) []byte {
	if kind == "alternative" {
		return m.ToColorAlternativeChangeSysEx()
	}
	var kind_byte byte = 0x06
	var sysex []byte
	sysex = append(sysex, 0x47, 0x7F, 0x4C, kind_byte, 0x00, 0x30)
	for i := 0; i < len(m.Elements); i++ {
		shouldLight := m.Elements[i].On || m.Elements[i].Static
		if m.Elements[i].Reverse {
			shouldLight = !shouldLight
		}
		if shouldLight {
			sysex = append(sysex, m.Elements[i].Color.ToAkaiBytes()...)
		} else {
			sysex = append(sysex, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
		}
	}
	return sysex
}

func (m *PadsStateArray) ToColorAlternativeChangeSysEx() []byte {
	var kind_byte byte = 0x06
	var sysex []byte
	sysex = append(sysex, 0x47, 0x7F, 0x4C, kind_byte, 0x03, 0x30)
	// for i := 0; i < len(m.Elements); i++ {
	// 	// sysex = append(sysex, m.Elements[i].AlternativeColor.ToAkaiBytes()...)
	// 	if i == 3 {
	// 		sysex = append(sysex, 127, 127, 127, 127, 127, 127)
	// 	} else {
	// 		sysex = append(sysex, 0, 0, 0, 0, 0, 0)
	// 	}

	// }
	// sysex = append(sysex, 0, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 0, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 0, 0, 0, 0, 0, 127)
	// // 3 - 3G 127 range
	// // 4 - 3G 127 range
	// // 5 - 3B 255 range
	// // 6 - 5R 127 range <<<
	// sysex = append(sysex, 0, 0, 0, 0)

	// // 1 - 5R 255 range
	// // 2 - 5G 255 range
	// // 3 - 5G 255 range
	// // 4 - 5B 127 range

	// sysex = append(sysex, 0, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 0, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 127, 0, 0, 0, 0, 0)
	// // sysex = append(sysex, 0, 0, 0, 0, 0, 0)

	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)
	sysex = append(sysex, 0, 0, 0)

	// sysex = append(sysex, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 0, 0, 0, 0, 0)

	// sysex = append(sysex, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 127, 0, 0, 0) //BB
	// sysex = append(sysex, 0, 0, 0, 0, 0, 0)
	// sysex = append(sysex, 0, 0, 0, 0, 0)

	return sysex
}
