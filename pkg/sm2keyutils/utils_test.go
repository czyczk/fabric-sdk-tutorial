package sm2keyutils

import (
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/XiaoYao-austin/ppks"
	"github.com/stretchr/testify/assert"
	"github.com/tjfoc/gmsm/x509"
)

func TestPrivKeyDER2PEMConversion(t *testing.T) {
	// Generate a private key. Convert it all the way to PEM and then back. Check if the products are as expected.
	privKey, err := ppks.GenPrivKey()
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	privKeyDer, err := x509.MarshalSm2UnecryptedPrivateKey(privKey)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	privKeyPemBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyDer,
	}

	privKeyPem := pem.EncodeToMemory(&privKeyPemBlock)
	fmt.Println(string(privKeyPem))

	decodedPrivKeyBlock, _ := pem.Decode(privKeyPem)
	if isEqual := assert.Equal(t, privKeyPemBlock.Bytes, decodedPrivKeyBlock.Bytes); !isEqual {
		t.FailNow()
	}

	unmarshalledPrivKey, err := x509.ParsePKCS8UnecryptedPrivateKey(decodedPrivKeyBlock.Bytes)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	if isEqual := assert.Equal(t, privKey, unmarshalledPrivKey); !isEqual {
		t.FailNow()
	}
}

func TestPubKeyDER2PEMConversion(t *testing.T) {
	// Generate a private key. Convert it all the way to PEM and then back. Check if the products are as expected.
	privKey, err := ppks.GenPrivKey()
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	pubKey := privKey.PublicKey

	pubKeyDer, err := x509.MarshalSm2PublicKey(&pubKey)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	pubKeyPemBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyDer,
	}

	pubKeyPem := pem.EncodeToMemory(&pubKeyPemBlock)
	fmt.Println(string(pubKeyPem))

	decodedPubKeyBlock, _ := pem.Decode(pubKeyPem)
	if isEqual := assert.Equal(t, pubKeyPemBlock.Bytes, decodedPubKeyBlock.Bytes); !isEqual {
		t.FailNow()
	}

	unmarshalledPubKey, err := x509.ParseSm2PublicKey(decodedPubKeyBlock.Bytes)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	if isEqual := assert.Equal(t, &pubKey, unmarshalledPubKey); !isEqual {
		t.FailNow()
	}
}
