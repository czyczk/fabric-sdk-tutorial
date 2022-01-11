package service

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/XiaoYao-austin/ppks"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
)

type integrityCheckResult int

const (
	matched integrityCheckResult = iota
	encryptedSizeNotMatched
	encryptedHashNotMatched
	decryptedSizeNotMatched
	decryptedHashNotMatched
)

func (r integrityCheckResult) toError(resourceEncryptionType data.ResourceType) error {
	switch r {
	case matched:
		return nil
	case encryptedSizeNotMatched:
		if resourceEncryptionType == data.Plain {
			panic("未知的错误类型")
		} else {
			return fmt.Errorf("获取的密文大小不正确")
		}
	case decryptedSizeNotMatched:
		if resourceEncryptionType == data.Plain {
			return fmt.Errorf("获取的资源大小不正确")
		} else {
			return fmt.Errorf("解密后的资源大小不正确")
		}
	case encryptedHashNotMatched:
		if resourceEncryptionType == data.Plain {
			panic("未知的错误类型")
		} else {
			return fmt.Errorf("获取的密文哈希不匹配")
		}
	case decryptedHashNotMatched:
		if resourceEncryptionType == data.Plain {
			return fmt.Errorf("获取的资源哈希不匹配")
		} else {
			return fmt.Errorf("解密后的资源哈希不匹配")
		}
	}

	panic("未知的检查结果类型")
}

func getResourceMetadata(id string, dataBCAO bcao.IDataBCAO) (*data.ResMetadataStored, error) {
	metadata, err := dataBCAO.GetMetadata(id)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// 同时从区块链的 CouchDB 和本地数据库的查询结果中挑取到满足分页数量要求的条目。
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

// 检查加密内容的大小和哈希是否匹配。
func checkSizeAndHashForEncryptedData(dataBytes []byte, metadata *data.ResMetadataStored) integrityCheckResult {
	encryptedSize := uint64(len(dataBytes))
	if encryptedSize != metadata.SizeStored {
		return encryptedSizeNotMatched
	}

	encryptedHash := sha256.Sum256(dataBytes)
	encryptedHashBase64 := base64.StdEncoding.EncodeToString(encryptedHash[:])
	if encryptedHashBase64 != metadata.HashStored {
		return encryptedHashNotMatched
	}

	return matched
}

// 检查明文或解密后的内容的大小和哈希是否匹配。
func checkSizeAndHashForDecryptedData(dataBytes []byte, metadata *data.ResMetadataStored) integrityCheckResult {
	size := uint64(len(dataBytes))
	if size != metadata.Size {
		return decryptedSizeNotMatched
	}

	hash := sha256.Sum256(dataBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	if hashBase64 != metadata.Hash {
		return decryptedHashNotMatched
	}

	return matched
}

// 解析并验证份额，若过程出现无法解析、无法验证或验证不通过的份额则返回错误，全部通过后解析的份额将通过列表返回。
func parseAndVerifySharesFromKeySwitchResults(ksResults []*keyswitch.KeySwitchResultStored, targetPublicKey *sm2.PublicKey, encryptedKey *ppks.CipherText, keySwitchService KeySwitchServiceInterface) ([]*ppks.CipherText, error) {
	var shares []*ppks.CipherText // 这里记录着通过了验证的份额
	for _, ksResult := range ksResults {
		shareBytes, err := base64.StdEncoding.DecodeString(ksResult.Share)
		if err != nil {
			return nil, errors.Wrap(err, "无法解析份额")
		}

		// 将每份 share 解析为一个 share *ppks.CipherText，并用对应的证明验证它。
		share, err := cipherutils.DeserializeCipherText(shareBytes)
		if err != nil {
			return nil, err
		}

		proofBytes, err := base64.StdEncoding.DecodeString(ksResult.ZKProof)
		if err != nil {
			return nil, errors.Wrap(err, "无法解析零知识证明")
		}

		proof, err := cipherutils.DeserializeZKProof(proofBytes)
		if err != nil {
			return nil, err
		}

		shareCreatorPublicKeyBytes, err := base64.StdEncoding.DecodeString(ksResult.KeySwitchPK)
		if err != nil {
			return nil, errors.Wrap(err, "无法解析份额创建者的密钥置换公钥")
		}

		shareCreatorPublicKey, err := cipherutils.DeserializeSM2PublicKey(shareCreatorPublicKeyBytes)
		if err != nil {
			return nil, err
		}

		isShareVerified, err := keySwitchService.VerifyShare(share, proof, shareCreatorPublicKey, targetPublicKey, encryptedKey)
		if err != nil {
			return nil, errors.Wrap(err, "无法验证份额")
		}

		if !isShareVerified {
			return nil, fmt.Errorf("份额验证未通过")
		}

		shares = append(shares, share)
	}

	return shares, nil
}
