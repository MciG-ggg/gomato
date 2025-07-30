package p2p

import (
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func loadOrCreatePrivateKey(path string) (crypto.PrivKey, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 生成新密钥
		privKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			return nil, err
		}

		// 保存
		keyData, err := crypto.MarshalPrivateKey(privKey)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, keyData, 0600); err != nil {
			return nil, err
		}
		fmt.Println("New key pair generated and saved.")
		return privKey, nil
	}
	// 读取已有密钥
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(keyData)
}
