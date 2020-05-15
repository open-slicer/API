package http

import (
	"github.com/gorilla/websocket"
	"net/http"
	"slicerapi/internal/util"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	util.Chk(err, true)
	if err != nil {
		return
	}

	for {
		t, msg, err := conn.ReadMessage()
		util.Chk(err, true)
		if err != nil {
			continue
		}
		// TODO: Handle messages.
		util.Chk(conn.WriteMessage(t, msg), true)
	}
}
