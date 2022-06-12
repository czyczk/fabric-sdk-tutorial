package fabricbcao

import (
	"encoding/hex"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/timingutils"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

func executeChannelRequestWithTimer(channelClient *channel.Client, channelRequest *channel.Request, timerMsg string) (resp channel.Response, err error) {
	defer timingutils.GetDeferrableTimingLogger(timerMsg)()

	resp, err = channelClient.Execute(*channelRequest)
	return
}

func getBlockHashFromTxID(ledgerClient *ledger.Client, txID fab.TransactionID) (string, error) {
	block, err := ledgerClient.QueryBlockByTxID(txID)
	if err != nil {
		return "", err
	}

	blockHashAsHex := hex.EncodeToString(block.GetHeader().GetDataHash())
	return blockHashAsHex, nil
}
