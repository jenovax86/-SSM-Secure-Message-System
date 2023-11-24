package ws

import "encoding/json"

type Event struct {
	Type    string          `json:"event"`
	Payload json.RawMessage `json:"data"`
}

type EventListener func(event *Event, c *Client) error

const (
	EventMessage = "message"
)

type MessageEvent struct {
	Message string `json:"message"`
	To      string `json:"to"`
}
