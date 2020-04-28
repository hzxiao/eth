package eth

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
)

// AddressFromPrivateKey 从密钥获得地址
func AddressFromPrivateKey(key string) (common.Address, error) {
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return common.Address{}, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("error casting public key to ECDSA")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}

// NewAddress 生成一个新地址，返回密钥，地址 十六进制表示
// 密钥返回对值没有'0x'作为前缀；地址则有
func NewAddress() (string, string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	privateKeyHex := hexutil.Encode(privateKeyBytes)[2:]
	address, err := AddressFromPrivateKey(privateKeyHex)
	if err != nil {
		return "", "", err
	}
	return privateKeyHex, address.Hex(), nil
}

// NewAddressByKeystore 通过keystore方式生成新地址
func NewAddressByKeystore(dir string, password string) (fp string, address string, err error) {
	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.NewAccount(password)
	if err != nil {
		return
	}

	address = account.Address.Hex()
	fp = account.URL.Path
	return
}

//ImportKeyStore 导入keystone文件生成KeyStore
func ImportKeyStore(tmpDir string, fp string, password string) (*keystore.KeyStore, string, error) {
	ks := keystore.NewKeyStore(tmpDir, keystore.StandardScryptN, keystore.StandardScryptP)
	jsonBytes, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, "", err
	}

	account, err := ks.Import(jsonBytes, password, password)
	if err != nil {
		return nil, "", err
	}

	return ks, account.URL.Path, nil
}
