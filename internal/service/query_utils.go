package service

import (
	"encoding/json"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

func getResourceMetadata(id string, serviceInfo *Info) (*data.ResMetadataStored, error) {
	chaincodeFcn := "getMetadata"
	channelReq := channel.Request{
		ChaincodeID: serviceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := serviceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, GetClassifiedError(chaincodeFcn, err)
	} else {
		var resMetadataStored data.ResMetadataStored
		if err = json.Unmarshal(resp.Payload, &resMetadataStored); err != nil {
			return nil, errors.Wrap(err, "获取的元数据不合法")
		}
		return &resMetadataStored, nil
	}
}

func collectResourceIDsFromSources(chaincodeResourceIDs, localDBResourceIDs []string, pageSize int, isReverse bool) (resourceIDs []string) {
	// 为两个数据源分别维护一个 `consumed` 变量，每从一个数据源中采用一条就将相应变量 +1。
	chaincodeSrcConsumed := 0
	localSrcConsumed := 0

	// 从两个来源采纳条目进结果列表。当结果列表的条目数量足够，或者两个来源均被遍历完（意为不再有新内容），则停止这一过程
	for {
		isChaincodeSrcLeft := chaincodeSrcConsumed < len(chaincodeResourceIDs)
		isLocalSrcLeft := localSrcConsumed < len(localDBResourceIDs)
		if isChaincodeSrcLeft && isLocalSrcLeft {
			// 按用户需求选择下一个条目。因为结果可能重复，在加入时若遇重复项，则只增对应的 `consumed` 变量，而将结果舍弃。
			isChaincodeSrcNext := true
			if !isReverse && chaincodeResourceIDs[chaincodeSrcConsumed] > localDBResourceIDs[localSrcConsumed] {
				isChaincodeSrcNext = false
			} else if isReverse && chaincodeResourceIDs[chaincodeSrcConsumed] < localDBResourceIDs[localSrcConsumed] {
				isChaincodeSrcNext = false
			}

			if isChaincodeSrcNext {
				if len(resourceIDs) == 0 || chaincodeResourceIDs[chaincodeSrcConsumed] != resourceIDs[len(resourceIDs)-1] {
					resourceIDs = append(resourceIDs, chaincodeResourceIDs[chaincodeSrcConsumed])
				}
				chaincodeSrcConsumed++
			} else {
				if len(resourceIDs) == 0 || localDBResourceIDs[localSrcConsumed] != resourceIDs[len(resourceIDs)-1] {
					resourceIDs = append(resourceIDs, localDBResourceIDs[localSrcConsumed])
				}
				localSrcConsumed++
			}
		} else if isChaincodeSrcLeft {
			// 链码来源有剩余，而本地数据库来源用完了。
			// 进到这里说明结果列表没满，进而说明本地数据库本次取出不足 `pageSize` 个，本地数据库已遍历完毕。
			// 那剩下的条目靠从链码来源的就可以了。
			// 在加入时若遇重复项，则只增对应的 `consumed` 变量，而将结果舍弃。
			if len(resourceIDs) == 0 || chaincodeResourceIDs[chaincodeSrcConsumed] != resourceIDs[len(resourceIDs)-1] {
				resourceIDs = append(resourceIDs, chaincodeResourceIDs[chaincodeSrcConsumed])
			}
			chaincodeSrcConsumed++
		} else if isLocalSrcLeft {
			// 链码来源不再有新内容，本地来源有剩余，那剩下的份额用本地数据库的填充即可。
			// 在加入时若遇重复项，则只增对应的 `consumed` 变量，而将结果舍弃。
			if len(resourceIDs) == 0 || localDBResourceIDs[localSrcConsumed] != resourceIDs[len(resourceIDs)-1] {
				resourceIDs = append(resourceIDs, localDBResourceIDs[localSrcConsumed])
			}
			localSrcConsumed++
		} else {
			// 两个来源均使用完毕，结束循环
			break
		}

		// 结果列表达到 `pageSize` 个，则可结束循环
		if len(resourceIDs) == pageSize {
			break
		}
	}

	return
}
