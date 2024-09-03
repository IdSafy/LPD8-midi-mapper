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
	On      bool // State of the element (on or off)
	Color   RGB  // RGB color of the element
	Reverse bool
	Static  bool
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
			On:      false,
			Color:   RGB{pads_config[i].RGB[0], pads_config[i].RGB[1], pads_config[i].RGB[2]},
			Reverse: pads_config[i].Reverse,
			Static:  pads_config[i].Static,
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

func (m *PadsStateArray) ToColorChangeSysEx() []byte {
	var sysex []byte
	sysex = append(sysex, 0x47, 0x7F, 0x4C, 0x06, 0x00, 0x30)
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
