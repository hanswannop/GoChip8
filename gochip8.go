// GoChip8, a Chip8 emulator written in go
// 2014 Hans Wannop

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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
	opcode     uint16     // Current opcode
	memory     [4096]byte // 4k of memory
	v          [16]byte   // Registers v[15] holds carry flag
	stack      [16]uint16 // Stack can hold 16 addresses for functions / callbacks
	sp         byte       // Stack pointer, index to top of stack
	index      uint16     // Temp store for addresses, only low 12 bits used
	pc         uint16     // Program counter (address of next instruction)
	delayTimer int        // Counts down at 60Hz when set > 0
	soundTimer int        // Counts down at 60Hz when set > 0
	width      int        // Screen width in pixels
	height     int        // Screen height in pixels
}

// Takes filename of the rom to load and returns initialised Chip8
func NewChip8(fileName string) *Chip8 {
	cpu := new(Chip8)
	cpu.pc = 0x200
	cpu.width = 64
	cpu.height = 32
	for i := 0; i < 80; i++ { // Load the font into the first 80 bytes of memory
		cpu.memory[i] = font[i]
	}
	rom, error := ioutil.ReadFile(fileName) //Open rom file
	if error != nil {
		panic(error)
	}
	// Need check here to make sure rom fits in memory
	for i := 0; i < len(rom); i++ { // Read rom into memory
		cpu.memory[0x200+i] = rom[i]
	}
	return cpu
}

//Step a cpu cycle
func (chip8 *Chip8) Step() {
	// Fetch
	chip8.opcode = uint16(chip8.memory[chip8.pc])<<8 | uint16(chip8.memory[chip8.pc+1]) // OR value of two consecutive addresses to get two byte opcode
	// Decode & execute
	switch chip8.opcode & 0xF000 {
	case 0x1000:
		{ // 1NNN Jumps to adress NNN
			chip8.pc = chip8.opcode & 0x0FFF
		}
	case 0x2000:
		{ // 2NNN Calls subroutine at NNN
			chip8.stack[chip8.sp] = chip8.pc
			chip8.sp++
			chip8.pc = chip8.opcode & 0x0FFF
		}
	case 0x3000:
		{ // 3XKK Skips next instruction if V[X] = KK
			if chip8.v[(chip8.opcode&0x0F00)>>8] == byte(chip8.opcode&0x00FF) {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
		}
	case 0x4000:
		{ // 4XKK Skips next instruction if V[X] != KK
			if chip8.v[(chip8.opcode&0x0F00)>>8] != byte(chip8.opcode&0x00FF) {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
		}
	case 0x5000:
		{ // 5XY0 Skips next instruction if V[X] == V[Y]
			if chip8.v[(chip8.opcode&0x0F00)>>8] == chip8.v[(chip8.opcode&0x00F0)>>4] {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
		}
	case 0x6000:
		{ // 6XKK Loads value KK into V[X]
			chip8.v[(chip8.opcode&0x0F00)>>8] = byte(chip8.opcode & 0x00FF)
			chip8.pc += 2
		}
	case 0xA000:
		{ // ANNN Sets index to address NNN
			chip8.index = chip8.opcode & 0x0FFF
			chip8.pc += 2
		}
	default:
		{
			fmt.Printf("Opcode not implemented: 0x%X\n", chip8.opcode)
			chip8.pc += 2
		}
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
	args := os.Args
	if len(args) > 1 {
		//fmt.Print(args[1])
		chip8 := NewChip8(args[1]) // Assume args[1] is filename of rom
		// 	for {  main loop off for developemnt
		chip8.Step() // Step cpu cycle
		// Should have check for draw flag here
		// Draw does not occur every cycle
		fmt.Print("\n", chip8)       // Refresh screen.
		time.Sleep(time.Second / 60) //Run at 60Hz
		//	}
		for i := 0; i < len(chip8.memory); i++ { // Print memory map to check things went well, for debug
			fmt.Printf("%X ", chip8.memory[i])
		}
		fmt.Print("\n")
	} else {
		fmt.Print("Must provide rom as argument.\n")
	}
}
