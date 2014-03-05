// GoChip8, a Chip8 emulator written in go
// 2014 Hans Wannop

package main

import (
	"bytes"
	"fmt"
	"time"
)

var font = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Chip8 struct {
	opcode     uint16
	memory     [4096]byte
	v          [16]byte
	stack      [16]uint16
	index      uint16
	pc         uint16
	delayTimer int
	soundTimer int
	width      int
	height     int
}

func NewChip8() *Chip8 {
	cpu := new(Chip8)
	cpu.pc = 0x200
	cpu.width = 64
	cpu.height = 32
	for i := 0; i < 80; i++ { // Load the font into the first 80 bytes of memory
		cpu.memory[i] = font[i]
	}
	return cpu
}

//Step a cpu cycle
func (chip *Chip8) Step() {
	// Fetch
	chip.opcode = uint16(chip.memory[chip.pc]<<8) | uint16(chip.memory[chip.pc+1]) // OR value of two consecutive addresses to get two byte opcode
	// Decode
	switch {
	// Execute
	}

}

// Renders the screen of the as a string.
func (chip8 *Chip8) String() string {
	var screenBuf bytes.Buffer
	for y := 0; y < chip8.height; y++ {
		for x := 0; x < chip8.width; x++ {
			b := byte(' ')
			//if associated pos in memory is 1 TODO
			//use a splice for screen?
			b = '*'
			screenBuf.WriteByte(b)
		}
		screenBuf.WriteByte('\n')
	}
	return screenBuf.String()
}

func main() {
	chip8 := NewChip8()
	for {
		chip8.Step() // Step cpu cycle
		// Should have check for draw flag here
		// Draw does not occur every cycle
		fmt.Print("\x0c", chip8)     // Refresh screen.
		time.Sleep(time.Second / 60) //Run at 60Hz
	}
}
