package ws

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"
	"golang.org/x/exp/slices"
)

type Channel struct {
	hub         *Hub
	Name        string
	Subscribers []*Client
}

func NewChanel(hub *Hub, name string) *Channel {
	result := &Channel{
		hub:         hub,
		Name:        name,
		Subscribers: []*Client{},
	}
	hub.channels = append(hub.channels, result)

	return result
}

func (ch *Channel) Unsubcribe(clientId string) {
	idx := slices.IndexFunc(ch.Subscribers, func(e *Client) bool { return e.Id == clientId })
	if idx != -1 {
		ch.Subscribers = append(ch.Subscribers[:idx], ch.Subscribers[idx+1:]...)
	}

	count := len(ch.Subscribers)
	if count == 0 {
		hub := ch.hub
		idxC := slices.IndexFunc(hub.channels, func(e *Channel) bool { return e.Name == ch.Name })
		if idxC != -1 {
			hub.channels = append(hub.channels[:idxC], hub.channels[idxC+1:]...)
		}
	}
}

// Subscribe a client to hub channels
func (hub *Hub) Subscribe(clientId string, channel string) bool {
	idx := slices.IndexFunc(hub.clients, func(c *Client) bool { return c.Id == clientId })

	if idx != -1 {
		client := hub.clients[idx]
		client.Subscribe(channel)

		event.Action("ws/subscribe", et.Json{"hub": hub.Id, "client": client, "channel": channel})

		return true
	}

	return false
}

func (hub *Hub) Unsubscribe(clientId string, channel string) bool {
	idx := slices.IndexFunc(hub.clients, func(c *Client) bool { return c.Id == clientId })

	if idx != -1 {
		client := hub.clients[idx]
		client.Unsubscribe(channel)

		event.Action("ws/unsubscribe", et.Json{"hub": hub.Id, "client": client, "channel": channel})

		return true
	}

	return false
}

func (hub *Hub) GetSubscribers(channel string) []*Client {
	idx := slices.IndexFunc(hub.channels, func(c *Channel) bool { return c.Name == channel })
	if idx != -1 {
		_channel := hub.channels[idx]
		return _channel.Subscribers
	}

	return []*Client{}
}
