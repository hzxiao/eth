package eth

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

// Signer 签名者
type Signer interface {
	SignTx(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
	CheckOwner(address string) error
}

// PKSigner 密钥签名
type PKSigner struct {
	pk *ecdsa.PrivateKey
}

// NewPKSigner 创建私钥签名者
func NewPKSigner(privateKey string) (*PKSigner, error) {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &PKSigner{pk: pk}, nil
}

// SignTx 对交易进行签名
func (signer *PKSigner) SignTx(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), signer.pk)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

// CheckOwner 检测拥有者
func (signer *PKSigner) CheckOwner(address string) error {
	publicKey := signer.pk.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	if crypto.PubkeyToAddress(*publicKeyECDSA).Hex() != address {
		return fmt.Errorf("address is not matched private key")
	}
	return nil
}

// KeyStoreSigner keystore签名者
type KeyStoreSigner struct {
	account accounts.Account
	ks      *keystore.KeyStore
}

// NewKeyStoreSigner 创建keystone签名者
func NewKeyStoreSigner(keydir string, address, password string) (*KeyStoreSigner, error) {
	ks := keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.Find(accounts.Account{Address: common.HexToAddress(address)})
	if err != nil {
		return nil, err
	}

	err = ks.Unlock(account, password)
	if err != nil {
		return nil, err
	}

	return &KeyStoreSigner{
		account: account,
		ks:      ks,
	}, nil
}

// SignTx 对交易进行签名
func (signer *KeyStoreSigner) SignTx(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	return signer.ks.SignTx(signer.account, tx, chainID)
}

// CheckOwner 检测拥有者
func (signer *KeyStoreSigner) CheckOwner(address string) error {
	if bytes.Compare(signer.account.Address.Bytes(), common.FromHex(address)) != 0 {
		return fmt.Errorf("address is not matched keystore")
	}
	return nil
}

// TransferTx 交易
type TransferTx struct {
	From  string
	To    string
	Value *big.Int
	//资产ID，以太坊为ETH，ERC20的token为对应的合约地址
	Asset string
	Hash  string
}

// ParseTransferETH 解析eth转账的交易，返回nil表示不是eth转账
func ParseTransferETH(tx *types.Transaction) (*TransferTx, error) {
	if tx == nil {
		return nil, fmt.Errorf("nil transaction")
	}

	if tx.Value().Cmp(big.NewInt(0)) == 0 || len(tx.Data()) > 0 || tx.To() != nil {
		return nil, nil
	}

	msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId()))
	if err != nil {
		return nil, err
	}

	ttx := &TransferTx{
		From:  msg.From().Hex(),
		To:    tx.To().Hex(),
		Value: tx.Value(),
		Hash:  tx.Hash().Hex(),
		Asset: "ETH",
	}
	return ttx, nil
}

// ParseTransferToken 解析转账token的交易
func ParseTransferToken(lg types.Log, tokens []string) (*TransferTx, error) {
	if lg.Removed {
		return nil, nil
	}

	if len(tokens) > 0 && !SliceContains(tokens, lg.Address.Hex()) {
		return nil, nil
	}

	if len(lg.Topics) != 3 {
		return nil, nil
	}
	signature := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	if lg.Topics[0].Hex() != signature.Hex() {
		return nil, nil
	}

	ttx := &TransferTx{
		From:  lg.Topics[1].Hex(),
		To:    lg.Topics[2].Hex(),
		Value: (&big.Int{}).SetBytes(lg.Data),
		Hash:  lg.TxHash.Hex(),
		Asset: lg.Address.Hex(),
	}
	return ttx, nil
}
