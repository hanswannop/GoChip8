// GoChip8, a Chip8 emulator written in go
// 2014 Hans Wannop

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
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
	opcode       uint16     // Current opcode
	memory       [4096]byte // 4k of memory
	v            [16]byte   // Registers v[15] holds carry flag
	stack        [16]uint16 // Stack can hold 16 addresses for functions / callbacks
	sp           byte       // Stack pointer, index to top of stack
	index        uint16     // Temp store for addresses, only low 12 bits used
	pc           uint16     // Program counter (address of next instruction)
	delayTimer   byte       // Counts down at 60Hz when set > 0
	soundTimer   byte       // Counts down at 60Hz when set > 0
	screen       []byte     // Each pos in slice is one pixel on screen
	width        uint16     // Screen width in pixels
	height       uint16     // Screen height in pixels
	needsDisplay bool       // Display flag, set for redraw
}

// TODO Returns true if key with specified code is pressed
func (chip8 *Chip8) KeyPressed(keycode byte) bool {
	return true
}

func NewChip8() *Chip8 {
	cpu := new(Chip8)
	cpu.pc = 0x200
	cpu.width = 64
	cpu.height = 32
	rand.Seed(time.Now().UTC().UnixNano()) // Seed random number generator
	cpu.screen = make([]byte, cpu.width*cpu.height) // Initialize the screen slice
	for i := 0; i < 80; i++ {                       // Load the font into the first 80 bytes of memory
		cpu.memory[i] = font[i]
	}
	return cpu
}

func (chip8 *Chip8) LoadRom(fileName string) {
	rom, error := ioutil.ReadFile(fileName) //Open rom file
	if error != nil {
		panic(error)
	}
	// Need check here to make sure rom fits in memory
	for i := 0; i < len(rom); i++ { // Read rom into memory
		chip8.memory[0x200+i] = rom[i]
	}
}

//Step a cpu cycle
func (chip8 *Chip8) Step() {
	// Fetch
	chip8.opcode = uint16(chip8.memory[chip8.pc])<<8 | uint16(chip8.memory[chip8.pc+1]) // OR value of two consecutive addresses to get two byte opcode
	// Decode & execute
	switch chip8.opcode & 0xF000 {
	case 0x0000:
		{
			switch chip8.opcode & 0x000F {
			case 0x0000: // 00E0 Clear the screen
				{
					for i := 0; i < len(chip8.screen); i++ {
						chip8.screen[i] = 0
					}
				}
			case 0x000E: // 000E Returns from subroutine
				{
					chip8.sp--
					chip8.pc = chip8.stack[chip8.sp]
					chip8.pc += 2
				}
			default:
				{
					fmt.Printf("Opcode not implemented: 0x%X\n", chip8.opcode)
					chip8.pc += 2
				}
			}

		}
	case 0x1000:
		{ // 1NNN Jumps to adress NNN
			chip8.pc = chip8.opcode & 0x0FFF
		}
	case 0x2000: // 2NNN Calls subroutine at NNN
		{
			chip8.stack[chip8.sp] = chip8.pc
			chip8.sp++
			chip8.pc = chip8.opcode & 0x0FFF
		}
	case 0x3000: // 3XKK Skips next instruction if V[X] = KK
		{
			if chip8.v[(chip8.opcode&0x0F00)>>8] == byte(chip8.opcode&0x00FF) {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
		}
	case 0x4000: // 4XKK Skips next instruction if V[X] != KK
		{
			if chip8.v[(chip8.opcode&0x0F00)>>8] != byte(chip8.opcode&0x00FF) {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
		}
	case 0x5000: // 5XY0 Skips next instruction if V[X] == V[Y]
		{
			if chip8.v[(chip8.opcode&0x0F00)>>8] == chip8.v[(chip8.opcode&0x00F0)>>4] {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
		}
	case 0x6000: // 6XKK Loads value KK into V[X]
		{
			chip8.v[(chip8.opcode&0x0F00)>>8] = byte(chip8.opcode & 0x00FF)
			chip8.pc += 2
		}
	case 0x7000: // 7XKK Adds KK to value stored in V[x]
		{
			chip8.v[(chip8.opcode&0x0F00)>>8] += byte(chip8.opcode & 0x00FF)
			chip8.pc += 2
		}
	case 0x8000: // Register binary arithmetic instuructions
		{
			switch chip8.opcode & 0x000F {
			case 0x000: // 8XY0 Loads value stored in V[Y] into V[X]
				{
					chip8.v[(chip8.opcode&0x0F00)>>8] = chip8.v[(chip8.opcode&0x00F0)>>4]
					chip8.pc += 2
				}
			case 0x001: // 8XY1 Bitwise OR between V[X] and V[Y], result stored in V[X]
				{
					chip8.v[(chip8.opcode&0x0F00)>>8] |= chip8.v[(chip8.opcode&0x00F0)>>4]
					chip8.pc += 2
				}
			case 0x002: // 8XY2 Bitwise AND between V[X] and V[Y], result stored in V[X]
				{
					chip8.v[(chip8.opcode&0x0F00)>>8] &= chip8.v[(chip8.opcode&0x00F0)>>4]
					chip8.pc += 2
				}
			case 0x003: // 8XY3 Bitwise XOR between V[X] and V[Y], result stored in V[X]
				{
					chip8.v[(chip8.opcode&0x0F00)>>8] ^= chip8.v[(chip8.opcode&0x00F0)>>4]
					chip8.pc += 2
				}
			case 0x004: // 8XY4 V[X] + V[Y], result stored in V[X], carry in V[0xF]
				{
					result := uint16(chip8.v[(chip8.opcode&0x0F00)>>8]) + uint16(chip8.v[(chip8.opcode&0x00F0)>>4])
					chip8.v[(chip8.opcode&0x0F00)>>8] = byte(result) //Need to check casting behaviour in go to confirm
					if result>>8 == 1 {
						chip8.v[0xF] = 1
					} else {
						chip8.v[0xF] = 0
					}
					chip8.pc += 2
				}
			case 0x005: // 8XY5 V[X] - V[Y], result stored in V[X], !borrow in V[0xF]
				{
					if chip8.v[(chip8.opcode&0x0F00)>>8] > chip8.v[(chip8.opcode&0x00F0)>>4] {
						chip8.v[0xF] = 1
					} else {
						chip8.v[0xF] = 0
					}
					chip8.v[(chip8.opcode&0x0F00)>>8] -= chip8.v[(chip8.opcode&0x00F0)>>4]
					chip8.pc += 2
				}
			case 0x006: // 8X06 Shift right (divide by 2), set V[0xF] to 1 if least significant bit was 1
				{
					if chip8.v[(chip8.opcode&0x0F00)>>8]&0x01 == 1 {
						chip8.v[0xF] = 1
					} else {
						chip8.v[0xF] = 0
					}
					chip8.v[(chip8.opcode&0x0F00)>>8] = chip8.v[(chip8.opcode&0x0F00)>>8] >> 1
					chip8.pc += 2
				}
			case 0x007: // 8XY7 V[Y] - V[X], result stored in V[X], !borrow in V[0xF]
				{
					if chip8.v[(chip8.opcode&0x0F0)>>4] > chip8.v[(chip8.opcode&0x0F00)>>8] {
						chip8.v[0xF] = 1
					} else {
						chip8.v[0xF] = 0
					}
					chip8.v[(chip8.opcode&0x0F00)>>8] = chip8.v[(chip8.opcode&0x00F0)>>4] - chip8.v[(chip8.opcode&0x0F00)>>8]
					chip8.pc += 2
				}
			case 0x00E: // 8X0E Shift left (multiply by 2), set V[0xF] to 1 if most significant bit was 1
				{
					if chip8.v[(chip8.opcode&0x0F00)>>8]&0x80 == 1 {
						chip8.v[0xF] = 1
					} else {
						chip8.v[0xF] = 0
					}
					chip8.v[(chip8.opcode&0x0F00)>>8] = chip8.v[(chip8.opcode&0x0F00)>>8] << 1
					chip8.pc += 2
				}
			}
		}
	case 0x9000: // 9XY0 Skips next instruction if V[X] != V[Y]
		{
			if chip8.v[(chip8.opcode&0x0F00)>>8] != chip8.v[(chip8.opcode&0x00F0)>>4] {
				chip8.pc += 4
			} else {
				chip8.pc += 2
			}
		}
	case 0xA000: // ANNN Sets index to address NNN
		{
			chip8.index = chip8.opcode & 0x0FFF
			chip8.pc += 2
		}
	case 0xB000: // BNNN Jump to address NNN + V[0]
		{
			chip8.pc = (chip8.opcode & 0x0FFF) + uint16(chip8.v[0])
		}
	case 0xC000: // CXKK Set V[X] to a random byte AND KK
		{
			chip8.v[0] = byte(uint16(rand.Int()) & (chip8.opcode & 0x00FF))
			chip8.pc += 2
		}
	case 0xD000: // DXYN Draw sprite stored at address INDEX with pos (V[X], V[Y]), width 8 and height N
		{ // V[OxF] is set if and pixel collides during draw
			xPos := uint16(chip8.v[(chip8.opcode&0x0F00)>>8])
			yPos := uint16(chip8.v[(chip8.opcode&0x00F0)>>4])
			spriteHeight := chip8.opcode & 0x000F
			chip8.v[0xF] = 0
			for row := uint16(0); row < spriteHeight; row++ {
				rowData := chip8.memory[chip8.index+row]
				for col := uint16(0); col < 8; col++ {
					if (rowData & (0x80 >> col)) != 0 { // If this pixel (row,col) in sprite in on
						if chip8.screen[((yPos+row)*chip8.width)+xPos+col] == 1 { //If pixel (xPos+col,yPos+row) in screen is on
							chip8.v[0xF] = 1
						}
						chip8.screen[((yPos+row)*chip8.width)+xPos+col] ^= 1
					}

				}
			}
			chip8.needsDisplay = true
			chip8.pc += 2
		}
	case 0xE000:
		{
			switch chip8.opcode & 0x00FF {
			case 0x009E: // EX9E Skip next instruction if key with value V[X] is pressed
				{
					if chip8.KeyPressed(chip8.v[(chip8.opcode&0x0F00)>>8]) {
						chip8.pc += 4
					} else {
						chip8.pc += 2
					}
				}
			case 0x00A1: // EXA1 Skip next instruction if key with value V[X] isn't pressed
				{
					if !chip8.KeyPressed(chip8.v[(chip8.opcode&0x0F00)>>8]) {
						chip8.pc += 4
					} else {
						chip8.pc += 2
					}
				}
			default:
				{
					fmt.Printf("Opcode not implemented: 0x%X\n", chip8.opcode)
					chip8.pc += 2
				}
			}
		}
	case 0xF000:
		{
			switch chip8.opcode & 0x00FF {
			case 0x0007: // FX07 Store delay timer value in V[X]
				{
					chip8.v[(chip8.opcode&0x0F00)>>8] = chip8.delayTimer
					chip8.pc += 2
				}
			case 0x000A: // FX07 Wait for key press, store key value in V[X] TODO
				{ // Need real wait for key press with actual keycode storage

					var input string
					fmt.Scanln(&input)
					chip8.v[(chip8.opcode&0x0F00)>>8] = 1
					chip8.pc += 2
				}
			case 0x0015: // FX15 Set delay timer to value of V[X]
				{
					chip8.delayTimer = chip8.v[(chip8.opcode&0x0F00)>>8]
					chip8.pc += 2
				}
			case 0x0018: // FX18 Set sound timer to value of V[X]
				{
					chip8.soundTimer = chip8.v[(chip8.opcode&0x0F00)>>8]
					chip8.pc += 2
				}
			case 0x001E: // FX1E Set INDEX = INDEX + V[X]
				{
					if chip8.index+uint16(chip8.v[(chip8.opcode&0x0F00)>>8]) > 0xFFF {
						chip8.v[0xF] = 1
					} else {
						chip8.v[0xF] = 0
					}
					chip8.index += uint16(chip8.v[(chip8.opcode&0x0F00)>>8])
					chip8.pc += 2
				}
			case 0x0029: // FX29 Set INDEX to memory address of sprite with number V[X]
				{
					chip8.index = uint16(chip8.v[(chip8.opcode&0x0F00)>>8] * 5) // Font characters are 5 bytes each
					chip8.pc += 2
				}
			case 0x0033: // FX33 Store decimal representation of V[X] in I, I+1, I+2
				{
					value := chip8.v[(chip8.opcode&0x0F00)>>8]
					chip8.memory[chip8.index] = value / 100
					chip8.memory[chip8.index+1] = value / 10 % 10
					chip8.memory[chip8.index+2] = (value % 100) % 10
					chip8.pc += 2
				}
			case 0x0055: // FX55 Store V[0] to V[X] in memory starting at address INDEX
				{
					lastReg := (chip8.opcode & 0x0F00) >> 8
					for i := uint16(0); i <= lastReg; i++ {
						chip8.memory[chip8.index+i] = chip8.v[i]
					}
					chip8.pc += 2
				}
			case 0x0065: // FX65 Load V[0] to V[X] from memory starting at address INDEX
				{
					lastReg := (chip8.opcode & 0x0F00) >> 8
					for i := uint16(0); i <= lastReg; i++ {
						chip8.v[i] = chip8.memory[chip8.index+i]
					}
					chip8.pc += 2
				}
			default:
				{
					fmt.Printf("Opcode not implemented: 0x%X\n", chip8.opcode)
					chip8.pc += 2
				}
			}
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
	for y := uint16(0); y < chip8.height; y++ {
		for x := uint16(0); x < chip8.width; x++ {
			var b byte
			if chip8.screen[y*chip8.width+x] != 0 {
				b = '*'
			} else {
				b = ' '
			}
			screenBuf.WriteByte(b)
		}
		screenBuf.WriteByte('\n')
	}
	return screenBuf.String()
}