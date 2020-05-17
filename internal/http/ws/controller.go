package ws

// Controller controls all websocket connections to the server.
type Controller struct {
	clients    map[string]*client
	broadcast  chan []byte
	register   chan *client
	unregister chan *client
}

// NewController creates a new controller instance.
func NewController() *Controller {
	return &Controller{
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[string]*client),
	}
}

// Run starts listening for channel inputs.
func (ctrl *Controller) Run() {
	for {
		select {
		case client := <-ctrl.register:
			ctrl.clients[client.username] = client
		case client := <-ctrl.unregister:
			if _, ok := ctrl.clients[client.username]; ok {
				close(client.send)
				delete(ctrl.clients, client.username)
			}
		case message := <-ctrl.broadcast:
			for _, client := range ctrl.clients {
				select {
				case client.send <- message:
				default:
					ctrl.unregister <- client
				}
			}
		}
	}
}
