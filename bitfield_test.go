package main

import (
	"math"
	"testing"
)

func TestGetPos(t *testing.T) {
	var size int32 = 128
	byteSize := int32(math.Ceil(float64(size) / 8.0))
	bitField := BitField{
		Size:     size,
		Bitfield: make([]byte, byteSize),
	}
	position := bitField.GetPos(100)
	if position.Byte != 12 {
		t.Errorf(
			"BitField: GetPos: Byte: size: %d expected: 12 actual: %d",
			size,
			position.Byte,
		)
	}

	if position.Bit != 4 {
		t.Errorf(
			"BitField: GetPos: Bit: size: %d expected: 4 actual: %d",
			size,
			position.Bit,
		)
	}
}

func TestOneAndUnset(t *testing.T) {
	var size int32 = 128
	byteSize := int32(math.Ceil(float64(size) / 8.0))
	bitField := BitField{
		Size:     size,
		Bitfield: make([]byte, byteSize),
	}
	bitField.One()
	for pos := int32(0); pos < size; pos++ {
		bit := bitField.GetBit(pos)
		if !bit {
			t.Errorf(
				"BitField: One: GetBit before: position: %d expected: %t actual: %t",
				pos,
				true,
				bit,
			)
		}
		bitField.UnsetBit(pos)
		bit = bitField.GetBit(pos)
		if bit {
			t.Errorf(
				"BitField: One: GetBit after: position: %d expected: %t actual: %t",
				pos,
				false,
				bit,
			)
		}
	}
}

func TestZeroAndSet(t *testing.T) {
	var size int32 = 128
	byteSize := int32(math.Ceil(float64(size) / 8.0))
	bitField := BitField{
		Size:     size,
		Bitfield: make([]byte, byteSize),
	}
	bitField.Zero()
	for pos := int32(0); pos < size; pos++ {
		bit := bitField.GetBit(pos)
		if bit {
			t.Errorf(
				"BitField: Zero: GetBit before: position: %d expected: %t actual: %t",
				pos,
				false,
				bit,
			)
		}
		bitField.SetBit(pos)
		bit = bitField.GetBit(pos)
		if !bit {
			t.Errorf(
				"BitField: Zero: GetBit after: position: %d expected: %t actual: %t",
				pos,
				true,
				bit,
			)
		}
	}
}
