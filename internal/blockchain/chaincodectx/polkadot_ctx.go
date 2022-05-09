package chaincodectx

import "gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"

type PolkadotChaincodeCtx struct {
	CallerAddress   string
	APIPrefix       string
	ContractAddress string
	ContractABI     string
}

func (ctx *PolkadotChaincodeCtx) GetBCType() blockchain.BCType {
	return blockchain.Polkadot
}
