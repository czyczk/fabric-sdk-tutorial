package service

import (
	"bytes"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/timingutils"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
)

func encryptDataWithTimer(bytes []byte, key *ppks.CurvePoint, errMsg string, timerMsg string) (encryptedBytes []byte, err error) {
	defer timingutils.GetDeferrableTimingLogger(timerMsg)()

	encryptedBytes, err = cipherutils.EncryptBytesUsingAESKey(bytes, cipherutils.DeriveSymmetricKeyBytesFromCurvePoint(key))
	if err != nil {
		err = errors.Wrap(err, errMsg)
		return
	}

	return
}

func executeChannelRequestWithTimer(channelClient *channel.Client, channelRequest *channel.Request, timerMsg string) (resp channel.Response, err error) {
	defer timingutils.GetDeferrableTimingLogger(timerMsg)()

	resp, err = channelClient.Execute(*channelRequest)
	return
}

func uploadBytesToIPFSWithTimer(ipfsSh *shell.Shell, encryptedBytes []byte, errMsg string, timerMsg string) (cid string, err error) {
	defer timingutils.GetDeferrableTimingLogger(timerMsg)()

	// Increase timeout for large files
	if len(encryptedBytes) > 1073741824 {
		ipfsSh.SetTimeout(120 * time.Second)
	} else {
		ipfsSh.SetTimeout(30 * time.Second)
	}

	cid, err = ipfsSh.Add(bytes.NewReader(encryptedBytes))
	if err != nil {
		err = errors.Wrap(err, "无法将加密后的文档上传至 IPFS 网络")
	}

	return
}
