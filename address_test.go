package eth

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestAddressFromPrivateKey(t *testing.T) {
	address, err := AddressFromPrivateKey("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	assert.NoError(t, err)
	t.Log(address.Hex())
	assert.Equal(t, "0x96216849c49358B10257cb55b28eA603c874b05E", address.Hex())
}

func TestNewAddress(t *testing.T) {
	pk, addr, err := NewAddress()
	assert.NoError(t, err)

	address, err := AddressFromPrivateKey(pk)
	assert.NoError(t, err)

	assert.Equal(t, addr, address.Hex())
}

func TestNewAddressByKeystore(t *testing.T) {
	base := "../../testdata/keystore/%v"
	for i := 0; i < 2; i++ {
		os.RemoveAll(fmt.Sprintf(base, i))

		fp, addr, err := NewAddressByKeystore(fmt.Sprintf(base, i), "123456")
		assert.NoError(t, err)

		ks, newFp, err := ImportKeyStore(os.TempDir(), fp, "123456")
		assert.NoError(t, err)

		assert.Len(t, ks.Accounts(), 1)
		assert.Equal(t, addr, ks.Accounts()[0].Address.Hex())
		os.Remove(newFp)
	}
}

func TestNewAddresses(t *testing.T) {
	for i := 0; i < 5; i++ {
		pk, addr, err := NewAddress()
		assert.NoError(t, err)

		t.Log(pk, addr)
	}
}
