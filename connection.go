package main

import (
    "github.com/gorilla/websocket"
    "net/http"
	"log"
)

type Connection struct {
    socket *websocket.Conn
	chip8 *Chip8 // currently each connection gets its own chip8 instance
}
var Connections = make([]*Connection, 0)

func wsHandler(w http.ResponseWriter, r *http.Request) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); ok {
        http.Error(w, "Not a websocket handshake", 400)
        return
    } else if err != nil {
        return
    }
	
	log.Printf("Websocket connection from %v\n", ws.RemoteAddr())
	
	connection := new(Connection)
	connection.socket = ws
	connection.chip8 = NewChip8() 
	Connections = append(Connections, connection)
}