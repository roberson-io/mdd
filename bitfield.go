package main

import (
	"math"
)

// Position represents a bit's location within the byte slice.
type Position struct {
	Byte uint
	Bit  uint
}

// BitField Implements bitfields for Bloom filters
type BitField struct {
	Size     int
	Bitfield []byte
	Position Position
}

// GetPos gets the position of a bit in a bitfield.
// Example:
// I want to get the position of the 100th bit. Since this is stored
// in a bytearray, one can't just do something like this:
//     value = bitfield[100] // This will get the 100th byte, not bit!
// The 100th bit of a bytearray will be 4 bits into the 12th byte:
//     >>> bitfield.getpos(100)
//     Position(byte=12, bit=4)
func (bf BitField) GetPos(position int) Position {
	var bytepos = uint(math.Ceil(float64(position)/8.0)) - 1
	var bitpos = uint(position % 8)
	if bitpos != 0 {
		bitpos = 8 - bitpos
	}
	return Position{
		Byte: bytepos,
		Bit:  bitpos,
	}
}

// SetBit sets the bit at specified position to 1.
func (bf BitField) SetBit(position int) {
	pos := bf.GetPos(position)
	bf.Bitfield[pos.Byte] |= (0x01 << pos.Bit) & 0xff
}

// UnsetBit sets the bit at specified position to 0.
func (bf BitField) UnsetBit(position int) {
	pos := bf.GetPos(position)
	bf.Bitfield[pos.Byte] |= ^(0x01 << pos.Bit) & 0xff
}

// GetBit retrieves the contents of a bit at a specific location.
func (bf BitField) GetBit(position int) bool {
	pos := bf.GetPos(position)
	contents := bf.Bitfield[pos.Byte] & ((0x01 << pos.Bit) & 0xff)
	return !(contents == 0)
}

// Zero sets all bits to zero.
func (bf BitField) Zero() {
	for _, position := range bf.Bitfield {
		bf.Bitfield[position] = 0x00
	}
}

// One sets all bits to one.
func (bf BitField) One() {
	for _, position := range bf.Bitfield {
		bf.Bitfield[position] = 0xff
	}
}
