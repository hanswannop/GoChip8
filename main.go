// GoChip8, a Chip8 emulator written in go
// 2014 Hans Wannop

package main

import (
  	"net/http"
 	"github.com/gorilla/mux"
	"log"
	"flag"
	"os"
	"fmt"
	"time"
)

var addr = flag.String("address", ":8080", "http address")
// Command-line flag to set (old) terminal mode, rather than running in browser 
var terminalMode = flag.Bool("terminal", false, "Run in terminal mode")

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method nod allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "html/index.html")
}

func main() {
	flag.Parse()
	
	if *terminalMode {
		args := os.Args
		if len(args) > 2 {
			fmt.Print(args[2])
			chip8 := NewChip8()             
			chip8.LoadRom(args[2]) // Require args[2] as filename of rom when in terminal mode
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
		r := mux.NewRouter()
		r.HandleFunc("/ws", wsHandler)
		r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
		if err := http.ListenAndServe(*addr, r); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}
} 
