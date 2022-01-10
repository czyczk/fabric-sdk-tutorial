package chaincodectx

import "gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"

type IChaincodeCtx interface {
	GetBCType() blockchain.BCType
}
