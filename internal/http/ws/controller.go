package ws

// C is the main controller used by the app. It's initialised in Run.
var C *Controller

// Controller controls all websocket connections to the server.
type Controller struct {
	Clients    map[string]*Client
	Channels   map[string]*Channel
	Broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// NewController creates a new controller instance.
func NewController(main bool) *Controller {
	c := &Controller{
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		Channels:   make(map[string]*Channel),
	}

	if main {
		C = c
	}
	return c
}

// Run starts listening for channel inputs. This blocks.
func (ctrl *Controller) Run() {
	for {
		select {
		case client := <-ctrl.register:
			ctrl.Clients[client.ID] = client
		case client := <-ctrl.unregister:
			if _, ok := ctrl.Clients[client.ID]; ok {
				close(client.Send)
				delete(ctrl.Clients, client.ID)
			}
		case message := <-ctrl.Broadcast:
			for _, client := range ctrl.Clients {
				select {
				case client.Send <- message:
				default:
					ctrl.unregister <- client
				}
			}
		}
	}
}
