package service

import (
	"fmt"
	"log"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr"
)

const eventTimeout time.Duration = 20

// 上传加密资源时的事件名称
const encryptedResourceCreationEventName = "enc_res_creation"

// Receives an event from the event channel and prints the event contents.
func showEventResult(notifier <-chan eventmgr.IEvent, eventID string) error {
	select {
	case ccEvent := <-notifier:
		log.Printf("Chaincode event received: %v\n", ccEvent)
	case <-time.After(eventTimeout * time.Second):
		return fmt.Errorf("can't receive chaincode event with event ID '%v'", eventID)
	}

	return nil
}
