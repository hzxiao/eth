package eth

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hzxiao/goutil/assert"
	"math/big"
	"os"
	"strings"
	"testing"
)

var Addr string
var RopstenUSDT = "0x4D15bd9027c65B2988391C4536daA1467F21B036"

func init() {
	err := setEnv("./.env")
	if err != nil {
		panic(err)
	}

	Addr = os.Getenv("ADDR")
}

func setEnv(fp string) error {
	f, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)

	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid line: %v", line)
		}

		err = os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		return err
	}
	return nil
}

type nonceManager struct {
	data map[string]uint64
}

func (n *nonceManager) GetNonceAt(address string) (uint64, error) {
	v := n.data[address] + 5
	v++
	n.data[address] = v
	return v, nil
}

func getClent(t *testing.T, b BlockNumberManager) *Client {

	nm := &nonceManager{
		data: make(map[string]uint64),
	}
	c, err := NewClient(Addr, b, nm, nil)
	assert.NoError(t, err)
	return c
}

func TestClinet(t *testing.T) {
	c, err := NewClient(Addr, nil, nil, nil)
	assert.NoError(t, err)

	cid, err := c.ethClient.ChainID(context.Background())
	assert.NoError(t, err)

	b, err := c.ethClient.SuggestGasPrice(context.Background())

	t.Log(b.String())
	t.Log(cid)

	v := &big.Int{}
	v.SetString("5378111721549909", 10)
	t.Log(v.Int64())

	bl, err := c.ethClient.BalanceAt(context.Background(), common.HexToAddress("0x4573eebA2Ff20A2559F86c2caf0605b1714d13E5"), big.NewInt(7772387))
	tx, _, err := c.ethClient.TransactionByHash(context.Background(), common.HexToHash("0x242db75760a19659b16b86d0d7a2d1519be78bd02635dbe38ce1e82b04cee38a"))

	j, _ := tx.MarshalJSON()
	t.Log(string(j))
	t.Log(bl.String())

	if msg, err := tx.AsMessage(types.NewEIP155Signer(tx.ChainId())); err == nil {
		t.Log(msg.From().Hex())

	} else {
		t.Log(err.Error())
	}

}

func TestClinet_TransferToken(t *testing.T) {
	cli := getClent(t, nil)

	cli.gasPrice = big.NewInt(2 * 10e9)
	signer, err := NewPKSigner("fc6b18ada631f43cb9768d89711b063c5cd07ace8c399377a336cc16446401b5")
	assert.NoError(t, err)

	from := "0x6D1e55FDf3110918e75500DD6387f25F2B5faa5d"
	to := "0xCD2269Eca29206AC970069020697aFC8f2b9ff62"
	txid, err := cli.TransferToken(RopstenUSDT, from, to, 60000, big.NewInt(10e6), signer)
	assert.NoError(t, err)

	t.Log(txid)

	// txid, err = cli.TransferETH(from, to, 21000, big.NewInt(8*1e17), signer)
	// assert.NoError(t, err)

}

func TestClinet_BalanceToken(t *testing.T) {
	cli := getClent(t, nil)

	b, err := cli.BalanceToken(RopstenUSDT, "0x4573eebA2Ff20A2559F86c2caf0605b1714d13E5", nil)
	assert.NoError(t, err)

	t.Log(b.String())
}
