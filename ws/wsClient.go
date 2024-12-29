package ws

import (
	"net/url"
	"os"
	"os/signal"

	"github.com/cgalvisleon/elvis/console"
	"github.com/gorilla/websocket"
)

type ClientWS struct {
	conn *websocket.Conn
}

// Create a new client websocket connection
func Cliente(host string, reciveFn func(message []byte)) *ClientWS {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	if host == "" {
		host = "localhost:8080"
	}

	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}

	// Connect to the server
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil
	}

	done := make(chan struct{})

	// Rutina para leer mensajes del servidor
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				console.AlertF(`Error al leer mensaje:%v`, err)
				return
			}
			if reciveFn != nil {
				reciveFn(message)
			}
		}
	}()

	return &ClientWS{conn: c}
}

// Close the client websocket connection
func (s *ClientWS) Close() {
	s.conn.Close()
}

func (s *ClientWS) Write(message string) error {
	msg := []byte(message)
	err := s.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return err
	}

	return nil
}
