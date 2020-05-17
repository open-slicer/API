package ws

import (
	"slicerapi/internal/util"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

// TODO: Don't use util.Chk as much.
// TODO: Add useful errors.
// TODO: Exit on some errors.
// TODO: Split up this file into multiple.

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512

	evtAddMessage = "EVT_ADD_MESSAGE"
	reqAddMessage = "REQ_ADD_MESSAGE"
)

var methods = map[string]func(*client, wsMessage){
	reqAddMessage: handleAddMessage,
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// wsMessage is a general message sent to or received by the server over WS. It's *not* specifically a chat message.
type wsMessage struct {
	Method string                 `json:"method"`
	Data   map[string]interface{} `json:"data"`
}

type client struct {
	controller *Controller
	conn       *websocket.Conn
	send       chan []byte
	username   string
}

func (c *client) readPump() {
	defer func() {
		c.controller.unregister <- c
		util.Chk(c.conn.Close(), true)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	util.Chk(c.conn.SetReadDeadline(time.Now().Add(pongWait)), true)
	c.conn.SetPongHandler(func(string) error {
		util.Chk(c.conn.SetReadDeadline(time.Now().Add(pongWait)), true)
		return nil
	})

	for {
		message := wsMessage{}
		err := c.conn.ReadJSON(&message)
		util.Chk(err, true)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				util.Chk(err, true)
			}
			break
		}

		if message.Method == "" {
			c.send <- []byte("method (string): required")
			continue
		}

		mth, ok := methods[message.Method]
		if !ok {
			c.send <- []byte("Invalid method")
			continue
		}

		mth(c, message)
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		util.Chk(c.conn.Close(), true)
	}()
	for {
		select {
		case message, ok := <-c.send:
			util.Chk(c.conn.SetWriteDeadline(time.Now().Add(writeWait)), true)
			if !ok {
				util.Chk(c.conn.WriteMessage(websocket.CloseMessage, []byte{}), true)
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			util.Chk(err, true)
			if err != nil {
				return
			}
			_, err = w.Write(message)
			util.Chk(err, true)

			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err = w.Write(<-c.send)
				util.Chk(err, true)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			util.Chk(c.conn.SetWriteDeadline(time.Now().Add(writeWait)), true)
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Handle handles new websocket connections.
func Handle(controller *Controller, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	util.Chk(err, true)
	if err != nil {
		return
	}

	claims := jwt.ExtractClaims(c)
	username := claims["username"].(string)
	client := &client{controller: controller, conn: conn, send: make(chan []byte, 256), username: username}
	client.controller.register <- client

	go client.writePump()
	go client.readPump()
}
