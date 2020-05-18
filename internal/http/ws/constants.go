package ws

const (
	evt    = "EVT_"
	evtAdd = evt + "ADD_"

	// EvtAddMessage is emitted when a message is sent.
	EvtAddMessage = evtAdd + "MESSAGE"
	// EvtAddChannel is emitted when a channel is created.
	EvtAddChannel = evtAdd + "CHANNEL"
	// EvtAddInvite is emitted when a user is invited to a channel.
	EvtAddInvite = evtAdd + "INVITE"
)
