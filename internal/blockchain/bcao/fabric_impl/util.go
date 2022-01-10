package fabric

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/timingutils"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

func executeChannelRequestWithTimer(channelClient *channel.Client, channelRequest *channel.Request, timerMsg string) (resp channel.Response, err error) {
	defer timingutils.GetDeferrableTimingLogger(timerMsg)()

	resp, err = channelClient.Execute(*channelRequest)
	return
}
