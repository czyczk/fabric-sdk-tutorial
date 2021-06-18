package service

import (
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

const eventTimeout time.Duration = 20

// RegisterEvent registers an event using a channel client.
//
// Returns:
//   the registration (used to unregister the event)
//   the event channel
func RegisterEvent(client *event.Client, chaincodeID, eventID string) (fab.Registration, <-chan *fab.CCEvent, error) {
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