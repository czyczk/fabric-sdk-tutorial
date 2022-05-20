package polkadoteventmgr

type PolkadotEvent struct {
	EventID string `json:"eventId"`
	Message string `json:"message"`
}

func (r PolkadotEventRegistration) GetEventID() string {
	return r.eventID
}

func (e PolkadotEvent) GetEventName() string {
	return e.EventID
}

func (e PolkadotEvent) GetPayload() []byte {
	return []byte(e.Message)
}

type PolkadotEventRegistration struct {
	contractAddress string
	eventID         string
}
