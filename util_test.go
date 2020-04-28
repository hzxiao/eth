package eth

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hzxiao/goutil/assert"
	"testing"
)

func TestHashMethod(t *testing.T) {
	methodID := HashMethod("transfer(address,uint256)")
	t.Log(hexutil.Encode(methodID))

	assert.Equal(t, "0xa9059cbb", hexutil.Encode(methodID))

	methodID = HashMethod("balanceOf(address)")

	assert.Equal(t, "0x70a08231", hexutil.Encode(methodID))
}
