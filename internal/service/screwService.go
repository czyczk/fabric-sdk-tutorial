package service

import (
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

type ScrewService struct {
	serviceInfo *Info
}

// TransferAndShowEvent invokes the chaincode with "transfer" command with the arguments specified.
//
// Parameters:
//   the name of the source corporation
//   the name of the destination corporation
//   the amount of screws
//
// Returns:
//   the transaction ID
func (s *ScrewService) TransferAndShowEvent(source, destination string, amount uint) (string, error) {
	// Try to register the event with ID "eventTransfer". Unregister it on failure.
	eventID := "eventTransfer"
	reg, notifier, err := registerEvent(s.serviceInfo.ChannelClient, s.serviceInfo.ChaincodeID, eventID)
	if err != nil {
		return "", err
	} else {
		defer s.serviceInfo.ChannelClient.UnregisterChaincodeEvent(reg)
	}

	// Make a channel request to invoke the "transfer" command.
	channelReq := channel.Request{
		ChaincodeID: s.serviceInfo.ChaincodeID,
		Fcn:         "transfer",
		Args:        [][]byte{[]byte(source), []byte(destination), []byte(strconv.Itoa(int(amount)))},
	}

	resp, err := s.serviceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", err
	}

	// Show the event result.
	if err = showEventResult(notifier, eventID); err != nil {
		return "", err
	}

	return string(resp.TransactionID), nil
}

// Query invokes the chaincode with "query" command with the arguments specified.
//
// Parameters:
//   the name of the corporation to query
//
// Returns:
//   the response payload
func (s *ScrewService) Query(corporationName string) (string, error) {
	channelReq := channel.Request{
		ChaincodeID: s.serviceInfo.ChaincodeID,
		Fcn:         "query",
		Args:        [][]byte{[]byte(corporationName)},
	}

	resp, err := s.serviceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return "", err
	}

	return string(resp.Payload), nil
}
