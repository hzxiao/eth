package eth

import (
	"github.com/ethereum/go-ethereum/crypto"
)

// HashMethod 对智能合约对方法进行哈希
func HashMethod(methodSignature string) []byte {
	h := crypto.Keccak256Hash([]byte(methodSignature))

	methodID := h.Bytes()[:4]
	return methodID
}

// SliceContains 检测是否包含目标字符串
func SliceContains(s []string, t string) bool {
	for i := range s {
		if s[i] == t {
			return true
		}
	}
	return false
}
