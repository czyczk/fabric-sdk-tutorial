package appinit

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// KeySwitchKeyLocations records the paths to the keys required to perform the key switch process.
type KeySwitchKeyLocations struct {
	EncryptionAlgorithm  string `yaml:"encryptionAlgorithm"`  // The encryption algorithm used by the keys
	CollectivePrivateKey string `yaml:"collectivePrivateKey"` // The path to the collective private key
	CollectivePublicKey  string `yaml:"collectivePublicKey"`  // The path to the collective public key
	PrivateKey           string `yaml:"privateKey"`           // The path to the private key
	PublicKey            string `yaml:"publicKey"`            // The path to the public key
}

// LoadKeySwitchServerKeys loads the keys required to perform the key switch process from the paths specified in `locations`. The keys will be available as singletons in `global.KeySwitchKeys`.
//
// Parameters:
//   a `KeySwitchKeyLocations` object containing the paths to the keys required to perform the key switch process
func LoadKeySwitchServerKeys(locations *KeySwitchKeyLocations) error {
	log.Info("正在为当前用户读取密钥置换所需的密钥...")

	// Notice: Only SM2 keys are planned to serve the key switch process currently.
	if strings.ToLower(locations.EncryptionAlgorithm) != "sm2" {
		return fmt.Errorf("只有 SM2 密钥可用于密钥置换")
	}

	global.KeySwitchKeys.EncryptionAlgorithm = locations.EncryptionAlgorithm

	// Load and save the collective private key as a singleton
	if locations.CollectivePrivateKey != "" {
		collPrivKeyPem, err := ioutil.ReadFile(locations.CollectivePrivateKey)
		if err != nil {
			return errors.Wrap(err, "无法读取密钥置换所需的集合私钥")
		}

		collPrivKey, err := sm2keyutils.ConvertPEMToPrivateKey(collPrivKeyPem)
		if err != nil {
			return errors.Wrap(err, "无法解析密钥置换所需的集合私钥，可能不是合法的 SM2 私钥")
		}

		global.KeySwitchKeys.CollectivePrivateKey = collPrivKey
	}

	// Load and save the collective public key as a singleton
	collPubKeyPem, err := ioutil.ReadFile(locations.CollectivePublicKey)
	if err != nil {
		return errors.Wrap(err, "无法读取密钥置换所需的集合公钥")
	}

	collPubKey, err := sm2keyutils.ConvertPEMToPublicKey(collPubKeyPem)
	if err != nil {
		return errors.Wrap(err, "无法解析密钥置换所需的集合公钥，可能不是合法的 SM2 公钥")
	}

	global.KeySwitchKeys.CollectivePublicKey = collPubKey

	// Load and save the private key as a singleton
	if locations.PrivateKey != "" {
		privKeyPem, err := ioutil.ReadFile(locations.PrivateKey)
		if err != nil {
			return errors.Wrap(err, "无法读取密钥置换所需的私钥")
		}

		privKey, err := sm2keyutils.ConvertPEMToPrivateKey(privKeyPem)
		if err != nil {
			return errors.Wrap(err, "无法解析密钥置换所需的私钥")
		}

		global.KeySwitchKeys.PrivateKey = privKey
	}

	// Load and save the public key as a singleton
	if locations.PublicKey != "" {
		pubKeyPem, err := ioutil.ReadFile(locations.PublicKey)
		if err != nil {
			return errors.Wrap(err, "无法读取密钥置换所需的公钥")
		}

		pubKey, err := sm2keyutils.ConvertPEMToPublicKey(pubKeyPem)
		if err != nil {
			return errors.Wrap(err, "无法解析密钥置换所需的公钥")
		}

		global.KeySwitchKeys.PublicKey = pubKey
	}

	return nil
}
