package service

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"strings"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Info needed for a service to know which `chaincodeID` it's serving and which channel it's using.
type Info struct {
	ChaincodeID   string
	ChannelClient *channel.Client
	EventClient   *event.Client
	LedgerClient  *ledger.Client
}

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

// GetClassifiedError is a general error handler that converts some errors returned from the chaincode to the predefined errors.
func GetClassifiedError(chaincodeFcn string, err error) error {
	if err == nil {
		return nil
	} else if strings.HasSuffix(err.Error(), errorcode.CodeForbidden) {
		return errorcode.ErrorForbidden
	} else if strings.HasSuffix(err.Error(), errorcode.CodeNotFound) {
		return errorcode.ErrorNotFound
	} else if strings.HasSuffix(err.Error(), errorcode.CodeNotImplemented) {
		return errorcode.ErrorNotImplemented
	} else {
		return errors.Wrapf(err, "无法调用链码函数 '%v'", chaincodeFcn)
	}
}
