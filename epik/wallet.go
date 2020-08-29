package epik

import (
	"fmt"
	"strings"

	"github.com/EpiK-Protocol/go-epik/chain/wallet"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/crypto"
)

var _wallet *wallet.Wallet
var _keyStore *wallet.MemKeyStore

func init() {
	fmt.Println("init wallet")
	_keyStore = wallet.NewMemKeyStore()

	_wallet, _ = wallet.NewWallet(_keyStore)
}

//GenerateKey ...
func GenerateKey(t string) (addrStr string, err error) {
	var addr address.Address
	switch strings.ToLower(t) {
	case "bls":
		addr, err = _wallet.GenerateKey(crypto.SigTypeBLS)
	case "secp256k1":
		addr, err = _wallet.GenerateKey(crypto.SigTypeSecp256k1)
	default:
		return "", fmt.Errorf("SigType not suppot")
	}
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}
