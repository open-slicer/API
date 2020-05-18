package ws

const (
	mEvt      = "EVT_"
	evtAdd    = mEvt + "ADD_"
	evtChange = mEvt + "CNG_"

	evtChangeListen = evtChange + "LISTEN"
	// EvtAddMessage is emitted when a message is sent.
	EvtAddMessage = evtAdd + "MESSAGE"
	// EvtAddChannel is emitted when a channel is created.
	EvtAddChannel = evtAdd + "CHANNEL"
	// EvtAddInvite is emitted when a user is invited to a channel.
	EvtAddInvite = evtAdd + "INVITE"

	mReq      = "REQ_"
	reqChange = mReq + "CNG_"

	reqChangeListen = reqChange + "LISTEN"

	mErr      = "ERR_"
	serverErr = mErr + "S_"
	clientErr = mErr + "C_"

	errJSON            = serverErr + "JSON"
	errMissingArgument = clientErr + "MISSING_ARG"
	errInvalidArgument = clientErr + "INVALID_ARG"
)
