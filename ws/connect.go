package ws

func connect() {
	hub := NewHub()
	go hub.Run()

	conn = &Conn{
		hub: hub,
	}
}
