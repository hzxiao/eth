package eth

// BlockNumberManager 已同步区块高度管理
type BlockNumberManager interface {
	// NumberSynced 获取最新已处理的区块高度
	NumberSynced() (uint64, error)
	// UpdateNumber 更新已同步的区块高度
	UpdateNumber(num uint64) error
	// IncreaseNumber 递增已同步的区块高度
	IncreaseNumber() error
}

// NonceManager nonce管理
type NonceManager interface {
	GetNonceAt(address string) (uint64, error)
}

// TransferHandler 转账回调处理
type TransferHandler interface {
	OnTransfer(tx TransferTx) error
}
