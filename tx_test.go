package eth

import (
	"github.com/hzxiao/goutil/assert"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"os"
	"testing"
)

func TestSigner(t *testing.T) {
	chainID := big.NewInt(1)
	pk, addr, err := NewAddress()
	assert.NoError(t, err)

	//signer1
	pkSigner, err := NewPKSigner(pk)
	assert.NoError(t, err)

	//signer2
	account, _, err := privateKeyToKeyStore(pk, os.TempDir(), "123456")
	assert.NoError(t, err)
	assert.Equal(t, addr, account.Address.Hex())
	ksSigner, err := NewKeyStoreSigner("", account.URL.Path, "123456")
	assert.NoError(t, err)
	defer os.Remove(account.URL.Path)

	//tx
	tx := createTx()

	//
	signedTx1, err := pkSigner.SignTx(tx, chainID)
	assert.NoError(t, err)

	signedTx2, err := ksSigner.SignTx(tx, chainID)
	assert.NoError(t, err)

	assert.Equal(t, signedTx1.Hash().Hex(), signedTx2.Hash().Hex())

}

func privateKeyToKeyStore(pk string, dir string, password string) (accounts.Account, *keystore.KeyStore, error) {
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return accounts.Account{}, nil, err
	}

	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.ImportECDSA(privateKey, password)

	return account, ks, err
}

func createTx() *types.Transaction {
	_, addr, _ := NewAddress()
	return types.NewTransaction(1, common.HexToAddress(addr), big.NewInt(100000000), 21000, big.NewInt(200), nil)
}
