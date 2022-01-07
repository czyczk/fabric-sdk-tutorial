package bcao

import (
	"net/url"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/contractctx"
)

func fillPostFormBase(ctx *contractctx.PolkadotContractCtx, form *url.Values) {
	form.Add("abi", ctx.ContractABI)
	form.Add("contractAddress", ctx.ContractAddress)
	form.Add("signerAddress", ctx.CallerAddress)
}
