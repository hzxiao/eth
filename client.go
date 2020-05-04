package eth

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"sync/atomic"
)

// Tokens ERC20代币合约地址，用于监听
var Tokens []string

// Client Eth Client
type Client struct {
	// eth node address
	Addr         string
	bmManager    BlockNumberManager
	nonceManager NonceManager

	transferHandler TransferHandler

	chainID   *big.Int
	ethClient *ethclient.Client

	gasPrice *big.Int
	// 已同步的区块数
	numSynced uint64
	closeCh   chan bool
}

// NewClient 创建新客户端
func NewClient(addr string, bmManager BlockNumberManager, nonceManager NonceManager, transferHandler TransferHandler) (*Client, error) {
	client := &Client{
		Addr:            addr,
		bmManager:       bmManager,
		nonceManager:    nonceManager,
		transferHandler: transferHandler,
		closeCh:         make(chan bool),
	}
	var err error
	client.ethClient, err = ethclient.Dial(addr)
	if err != nil {
		return nil, err
	}

	client.chainID, err = client.ethClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	client.gasPrice, err = client.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// SetGasPrice 设置gas价格
func (c *Client) SetGasPrice(p *big.Int) {
	if p != nil {
		c.gasPrice = p
	}
}

// StartWatching 开始监听区块变化
func (c *Client) StartWatching() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic by cause(%v)", e)
		}
	}()
	c.numSynced, err = c.bmManager.NumberSynced()
	if err != nil {
		return err
	}

	headers := make(chan *types.Header)
	sub, err := c.ethClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case header := <-headers:

			var step uint64 = 5
			var lastest uint64
			lastest = header.Number.Uint64()
			for i := c.numSynced + 1; i <= lastest; i += step {
				i = c.numSynced + 1
				var high = i + step - 1
				if high > lastest {
					high = lastest
				}
				err = c.SyncNewBlocks(i, high)
				if err != nil {
					//TODO: handler
					return err
				}
			}
		case err := <-sub.Err():
			_ = err
		case <-c.closeCh:
			break
		}
	}
}

// SyncNewBlocks 同步新区块
func (c *Client) SyncNewBlocks(low, high uint64) error {
	if high <= atomic.LoadUint64(&c.numSynced) {
		return nil
	}
	query := ethereum.FilterQuery{
		FromBlock: (&big.Int{}).SetUint64(low),
		ToBlock:   (&big.Int{}).SetUint64(high),
		Topics:    [][]common.Hash{{crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))}},
	}
	for _, token := range Tokens {
		query.Addresses = append(query.Addresses, common.HexToAddress(token))
	}

	vLogs, err := c.ethClient.FilterLogs(context.Background(), query)
	if err != nil {
		return err
	}

	var mLog = make(map[uint64][]*types.Log)
	for i := range vLogs {
		mLog[vLogs[i].BlockNumber] = append(mLog[vLogs[i].BlockNumber], &vLogs[i])
	}
	vLogs = nil
	//block
	for i := low; i <= high; i++ {
		num := (&big.Int{}).SetUint64(i)
		block, err := c.ethClient.BlockByNumber(context.Background(), num)
		if err != nil {
			return err
		}
		for _, tx := range block.Transactions() {
			ttx, err := ParseTransferETH(tx)
			if err != nil {
				//TODO:
			}
			if ttx == nil {
				continue
			}
			err = c.transferHandler.OnTransfer(*ttx)
			if err != nil {
				return err
			}
		}
		for _, lg := range mLog[i] {
			ttx, _ := ParseTransferToken(*lg, Tokens)
			if ttx == nil {
				continue
			}
			err = c.transferHandler.OnTransfer(*ttx)
			if err != nil {
				return err
			}
		}
		// 更新已同步的区块数
		err = c.bmManager.IncreaseNumber()
		if err != nil {
			return err
		}
		atomic.AddUint64(&c.numSynced, 1)

	}
	return nil
}

// SyncBlock 同步指定的区块
func (c *Client) SyncBlock(num uint64) error {
	return nil
}

// TransferETH 转账ETH
func (c *Client) TransferETH(from, to string, gasLimit uint64, value *big.Int, signer Signer) (string, uint64, error) {
	return c.completeAndSendTx(from, to, gasLimit, value, nil, signer)
}

// TransferToken 转账token
func (c *Client) TransferToken(token, from, to string, gasLimit uint64, value *big.Int, signer Signer) (txid string, nonce uint64, err error) {
	methodID := HashMethod("transfer(address,uint256)")
	paddedAddress := common.LeftPadBytes(common.HexToAddress(to).Bytes(), 32)
	paddedAmount := common.LeftPadBytes(value.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	return c.completeAndSendTx(from, token, gasLimit, big.NewInt(0), data, signer)
}

func (c *Client) completeAndSendTx(from, to string, gasLimit uint64, value *big.Int, data []byte, signer Signer) (string, uint64, error) {
	err := signer.CheckOwner(from)
	if err != nil {
		return "", 0, err
	}
	nonce, err := c.nonceManager.GetNonceAt(from)
	if err != nil {
		return "", 0, err
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(to), value, gasLimit, c.gasPrice, data)

	signedTx, err := signer.SignTx(tx, c.chainID)
	if err != nil {
		return "", 0, err
	}

	err = c.ethClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", 0, err
	}

	return signedTx.Hash().Hex(), nonce, nil
}

// BalanceETH 查询指定地址的ETH数量，单位wei
// blockNumber 为nil时，为最新区块
func (c *Client) BalanceETH(address string, blockNumber *big.Int) (*big.Int, error) {
	return c.ethClient.BalanceAt(context.Background(), common.HexToAddress(address), blockNumber)
}

// BalanceToken 查询指定地址指定ERC20的token的数量
// blockNumber 为nil时，为最新区块
func (c *Client) BalanceToken(token, address string, blockNumber *big.Int) (*big.Int, error) {
	tokenAddress := common.HexToAddress(token)
	paddingAddress := common.LeftPadBytes(common.HexToAddress(address).Bytes(), 32)
	result, err := c.ethClient.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: append(HashMethod("balanceOf(address)"), paddingAddress...),
	}, blockNumber)
	if err != nil {
		return nil, err
	}
	return (&big.Int{}).SetBytes(result), nil
}

// Status 客户端状态信息
func (c *Client) Status() map[string]interface{} {
	return map[string]interface{}{}
}

// Close 关闭
func (c *Client) Close() error {
	c.ethClient.Close()
	close(c.closeCh)
	return nil
}
