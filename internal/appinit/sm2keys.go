package appinit

import (
	"io/ioutil"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// KeyPairLocation records the paths to a key pair.
type KeyPairLocation struct {
	PrivateKey string `yaml:"privateKey"` // The path to the private key
	PublicKey  string `yaml:"publicKey"`  // The path to the public key
}

// LoadSM2KeyPair loads a SM2 key pair from the paths specified in `location`. The keys will be available as singletons in `global.SM2PrivateKey` and `global.SM2PublicKey`.
//
// Parameters:
//   a key pair location object
func LoadSM2KeyPair(location *KeyPairLocation) error {
	log.Info("正在为当前用户读取 SM2 密钥对...")

	// Load and save the private key as a singleton
	privKeyPem, err := ioutil.ReadFile(location.PrivateKey)
	if err != nil {
		return errors.Wrapf(err, "无法读取 SM2 私钥")
	}

	privKey, err := sm2keyutils.ConvertPEMToPrivateKey(privKeyPem)
	if err != nil {
		return errors.Wrapf(err, "无法解析 SM2 私钥")
	}

	global.SM2PrivateKey = privKey

	// Load and save the public key as a singleton
	pubKeyPem, err := ioutil.ReadFile(location.PublicKey)
	if err != nil {
		return errors.Wrapf(err, "无法读取 SM2 公钥")
	}

	pubKey, err := sm2keyutils.ConvertPEMToPublicKey(pubKeyPem)
	if err != nil {
		return errors.Wrapf(err, "无法解析 SM2 公钥")
	}

	global.SM2PublicKey = pubKey

	return nil
}
