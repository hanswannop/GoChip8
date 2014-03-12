// GoChip8, a Chip8 emulator written in go
// 2014 Hans Wannop

package main

import (
//  "net/http"
//	"log"
	"flag"
	"os"
	"fmt"
	"time"
)

// Command-line flag to set (old) terminal mode, rather than running in browser 
var terminalMode = flag.Bool("terminal", false, "Run in terminal mode")

func main() {
	flag.Parse()

	if *terminalMode {
		args := os.Args
		if len(args) > 2 {
			fmt.Print(args[2])
			chip8 := NewChip8(args[2])             // Assume args[1] is filename of rom
			for {
				chip8.Step()              // Step cpu cycle
				if chip8.delayTimer > 0 { // Update in seperate thread to keep at 60Hz?
					chip8.delayTimer--
				}
				if chip8.needsDisplay {
					fmt.Print("\n", chip8) // Refresh screen.
					chip8.needsDisplay = false
				}
				//fmt.Printf("%X ", chip8.opcode)
				time.Sleep(time.Second / 1000) //Run at 60Hz
			}
			//for i := 0; i < len(chip8.memory); i++ { // Print memory map on exit
			//	fmt.Printf("%X ", chip8.memory[i])
			//}
			//fmt.Print("\n")
		} else {
			fmt.Print("Must provide rom as argument.\n")
		}	
	} else {
		// Serve web page
	}
} 
