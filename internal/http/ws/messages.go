package ws

import "slicerapi/internal/db"

// Message is a general message sent to or received by the server over WS. It's *not* specifically a chat message.
type Message struct {
	Method string `json:"method"`
}

// ErrMessage is a generic error message.
type ErrMessage struct {
	Message
	Data string `json:"data"`
}

// ChannelMessage is a ws message including a channel.
type ChannelMessage struct {
	Message
	Data db.Channel `json:"data"`
}

// ChatMessage is a message including a db.Message.
type ChatMessage struct {
	Message
	Data db.Message `json:"data"`
}
