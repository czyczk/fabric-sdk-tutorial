package main

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/XiaoYao-austin/ppks"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

func generateKeys(dirKeys string, users []string) error {
	// Exit if the dir exists
	if _, err := os.Stat(dirKeys); os.IsExist(err) {
		return fmt.Errorf("the sm2 keys are already generated. Delete the folder first before running again")
	}

	// Create the dir
	os.Mkdir(dirKeys, 0755)

	// Collect public keys to generate a collective public key
	pubKeys := []sm2.PublicKey{}

	for _, user := range users {
		// Generate keys
		privKey, err := ppks.GenPrivKey()
		if err != nil {
			return errors.Wrapf(err, "cannot generate a private key for '%v'", user)
		}

		pubKey := privKey.PublicKey
		pubKeys = append(pubKeys, pubKey)

		// Create a directory for the user
		if _, err = os.Stat(path.Join(dirKeys, user)); os.IsNotExist(err) {
			os.Mkdir(path.Join(dirKeys, user), 0755)
		}

		// Save the private key and the public key to files
		// Private key
		privKeyDer, err := x509.MarshalSm2UnecryptedPrivateKey(privKey)
		if err != nil {
			return errors.Wrapf(err, "cannot save the private key for '%v'", user)
		}
		privKeyPemBlock := pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privKeyDer,
		}
		privKeyPem := pem.EncodeToMemory(&privKeyPemBlock)
		ioutil.WriteFile(path.Join(dirKeys, user, "sk"), privKeyPem, 0644)

		// Public key
		pubKeyDer, err := x509.MarshalSm2PublicKey(&pubKey)
		if err != nil {
			return errors.Wrapf(err, "cannot save the public key for '%v'", user)
		}
		pubKeyPemBlock := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubKeyDer,
		}
		pubKeyPem := pem.EncodeToMemory(&pubKeyPemBlock)
		ioutil.WriteFile(path.Join(dirKeys, user, user+".pem"), pubKeyPem, 0644)
	}

	// Construct a collective public key and save it
	collPubKey := ppks.CollPubKey(pubKeys)
	collPubKeyDer, err := x509.MarshalSm2PublicKey(collPubKey)
	if err != nil {
		return errors.Wrap(err, "cannot save the collective public key")
	}
	collPubKeyPemBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: collPubKeyDer,
	}
	collPubKeyPem := pem.EncodeToMemory(&collPubKeyPemBlock)
	ioutil.WriteFile(path.Join(dirKeys, "collPubKey.pem"), collPubKeyPem, 0644)

	return nil
}
