package polkadot

// TxExecutionResult contains all the info to be returned to the client about the result of the transaction execution.
type TxExecutionResult struct {
	TxHash        string                `json:"txHash"`
	DispatchInfo  *polkadotDispatchInfo `json:"dispatchInfo"`
	InBlockStatus *inBlockStatus        `json:"inBlockStatus"`
}

// ContractInstantiationSuccessResult contains all the info to be returned to the client about the result of a successful contract instantiation.
type ContractInstantiationSuccessResult struct {
	TxExecutionResult
	Address string `json:"address"`
}

// ContractInstantiationErrorResult contains all the info to be returned to the client about the result of a failed contract instantiation.
type ContractInstantiationErrorResult struct {
	TxExecutionResult
	ExplainedModuleError *ExplainedModuleError `json:"explainedModuleError"`
}

// ContractQuerySuccessResult contains all the info to be returned to the client about the result of a successful contract query.
type ContractQuerySuccessResult struct {
	contractQueryResultBase
	Output string `json:"output"`
}

// ContractQueryErrorResult contains all the info to be returned to the client about the result of a failed contract query.
type ContractQueryErrorResult struct {
	contractQueryResultBase
	ExplainedModuleError *ExplainedModuleError `json:"explainedModuleError"`
	DebugMessage         string                `json:"debugMessage"`
}

// ContractTxSuccessResult contains all the info to be returned to the client about the result of a successful contract transaction.
type ContractTxSuccessResult struct {
	TxExecutionResult
	ParsedContractEvents *string `json:"parsedContractEvents"`
}

// ContractTxErrorResult contains all the info to be returned to the client about the result of a failed contract transaction.
type ContractTxErrorResult struct {
	TxExecutionResult
	ExplainedModuleError *ExplainedModuleError `json:"explainedModuleError"`
}

type polkadotDispatchInfo struct {
	Weight  int    `json:"weight"`
	Class   string `json:"class"`
	PaysFee string `json:"paysFee"`
}

type inBlockStatus struct {
	InBlock   string
	Finalized *string
}

type ExplainedModuleError struct {
	Index   byte   `json:"index"`
	Error   byte   `json:"error"`
	Type    string `json:"type"`
	Details string `json:"details"`
}

// Contains the info that all contract query results will include.
type contractQueryResultBase struct {
	GasConsumed int `json:"gasConsumed"`
}
