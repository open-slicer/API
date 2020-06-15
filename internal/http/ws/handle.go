package ws

import (
	"encoding/json"
	"net/http"
	"slicerapi/internal/util"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

// TODO: Don't use util.Chk as much.
// TODO: Add useful errors.
// TODO: Exit on some errors.

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_r *http.Request) bool {
		return true
	},
}

// Client is a websocket client interfacing with the server.
type Client struct {
	conn  *websocket.Conn
	index int
	Send  chan []byte
	ID    string
}

func (c *Client) readPump() {
	defer func() {
		C.unregister <- c
		util.Chk(c.conn.Close(), true)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	util.Chk(c.conn.SetReadDeadline(time.Now().Add(pongWait)), true)
	c.conn.SetPongHandler(func(string) error {
		util.Chk(c.conn.SetReadDeadline(time.Now().Add(pongWait)), true)
		return nil
	})

	for {
		// TODO: This **really** needs to be adapted to allow for multiple methods to be used.
		message := changeListenMessage{}
		err := c.conn.ReadJSON(&message)
		util.Chk(err, true)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				util.Chk(err, true)
			}
			break
		}

		// If a method isn't present, send ERR_C_MISSING_ARG.
		if message.Method == "" {
			marshalled, err := json.Marshal(ErrMessage{
				Message: Message{Method: errMissingArgument},
				Data:    "method",
			})
			if err != nil {
				util.Chk(err, true)
				c.Send <- []byte(errJSON)
			}

			c.Send <- marshalled
			continue
		}

		// Handle the REQ_CNG_LISTEN websocket request.
		if message.Method != reqChangeListen {
			marshalled, err := json.Marshal(ErrMessage{
				Message: Message{Method: errInvalidArgument},
				Data:    "method",
			})
			if err != nil {
				util.Chk(err, true)
				c.Send <- []byte(errJSON)
			}

			c.Send <- marshalled
			continue
		}

		handleChangeListen(c, message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		util.Chk(c.conn.Close(), true)
	}()
	for {
		select {
		case message, ok := <-c.Send:
			util.Chk(c.conn.SetWriteDeadline(time.Now().Add(writeWait)), true)
			if !ok {
				// Close if not ok.
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

			// Write the rest of the send channel if possible.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, err = w.Write(<-c.Send)
				util.Chk(err, true)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// Send the ping message on tick.
			util.Chk(c.conn.SetWriteDeadline(time.Now().Add(writeWait)), true)
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Handle handles new websocket connections.
func Handle(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	util.Chk(err, true)
	if err != nil {
		return
	}

	claims := jwt.ExtractClaims(c)
	id := claims["id"].(string)

	index := 0
	if clientConnections, ok := C.Clients[id]; ok {
		index = len(clientConnections)
	}
	client := &Client{conn: conn, Send: make(chan []byte, 256), ID: id, index: index}
	C.register <- client

	go client.writePump()
	go client.readPump()
}
