package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr"
)

// ScrewService contains functions that accept parameters valid in type and invoke the corresponding chaincode functions.
type ScrewService struct {
	ScrewBCAO    bcao.IScrewBCAO
	EventManager eventmgr.IEventManager
}

// TransferAndShowEvent invokes the chaincode with "transfer" command with the arguments specified.
//
// Parameters:
//   the name of the source corporation
//   the name of the target corporation
//   the amount of screws
//
// Returns:
//   the transaction ID
func (s *ScrewService) TransferAndShowEvent(source, target string, amount uint) (string, error) {
	// Try to register the event with ID "eventTransfer". Unregister it on failure.
	eventID := "eventTransfer"
	reg, notifier, err := s.EventManager.RegisterEvent(eventID)
	if err != nil {
		return "", err
	}

	defer s.EventManager.UnregisterEvent(reg)

	// Make a channel request to invoke the "transfer" command.
	txID, err := s.ScrewBCAO.Transfer(source, target, amount, eventID)
	if err != nil {
		return "", err
	}

	// Show the event result.
	if err = showEventResult(notifier, eventID); err != nil {
		return "", err
	}

	return txID, nil
}

// Query invokes the chaincode with "query" command with the arguments specified.
//
// Parameters:
//   the name of the corporation to query
//
// Returns:
//   the response payload. Empty payloads imply invalid query keys.
func (s *ScrewService) Query(corporationName string) (string, error) {
	return s.ScrewBCAO.Query(corporationName)
}
