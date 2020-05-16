package ws

type controller struct {
	clients    map[string]*client
	broadcast  chan []byte
	register   chan *client
	unregister chan *client
}

func NewController() *controller {
	return &controller{
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[string]*client),
	}
}

func (ctrl *controller) Run() {
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
