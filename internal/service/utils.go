package service

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	log "github.com/sirupsen/logrus"
)

// Info needed for a service to know which `chaincodeID` it's serving and which channel it's using.
type Info struct {
	ChaincodeID   string
	ChannelClient *channel.Client
}

const eventTimeout time.Duration = 20

// Registers an event with a channel client.
//
// Returns:
//   the registration (used to unregister the event)
//   the event channel
func registerEvent(client *channel.Client, chaincodeID, eventID string) (fab.Registration, <-chan *fab.CCEvent, error) {
	reg, notifier, err := client.RegisterChaincodeEvent(chaincodeID, eventID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register chaincode event: %v", err)
	}
	return reg, notifier, nil
}

// Receives an event from the event channel and prints the event contents.
func showEventResult(notifier <-chan *fab.CCEvent, eventID string) error {
	select {
	case ccEvent := <-notifier:
		log.Printf("Chaincode event received: %v\n", ccEvent)
	case <-time.After(eventTimeout * time.Second):
		return fmt.Errorf("can't receive chaincode event with event ID '%v'", eventID)
	}

	return nil
}
