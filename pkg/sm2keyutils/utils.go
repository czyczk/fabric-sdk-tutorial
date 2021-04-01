package sm2keyutils

import (
	"encoding/pem"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

// Convert a PEM formatted private key to an `sm2.PrivateKey` object.
func ConvertPEMToPrivateKey(pemBytes []byte) (*sm2.PrivateKey, error) {
	decodedPrivKeyBlock, _ := pem.Decode(pemBytes)

	parsedPrivKey, err := x509.ParsePKCS8UnecryptedPrivateKey(decodedPrivKeyBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert PEM to SM2 private key")
	}

	return parsedPrivKey, nil
}

// Convert an `sm2.PrivateKey` object to PEM formatted bytes.
func ConvertPrivateKeyToPEM(privKey *sm2.PrivateKey) ([]byte, error) {
	privKeyDer, err := x509.MarshalSm2UnecryptedPrivateKey(privKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert private key to PEM")
	}

	privKeyPemBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyDer,
	}

	privKeyPem := pem.EncodeToMemory(&privKeyPemBlock)
	return privKeyPem, nil
}

// Convert a big integer to an `sm2.PrivateKey` object.
func ConvertBigIntegerToPrivateKey(d *big.Int) *sm2.PrivateKey {
	c := sm2.P256Sm2()

	priv := new(sm2.PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = d
	priv.PublicKey.X, priv.PublicKey.Y = c.ScalarBaseMult(d.Bytes())

	return priv
}

// Convert a PEM formatted public key to an `sm2.PublicKey` object.
func ConvertPEMToPublicKey(pemBytes []byte) (*sm2.PublicKey, error) {
	decodedPubKeyBlock, _ := pem.Decode(pemBytes)

	parsedPubKey, err := x509.ParseSm2PublicKey(decodedPubKeyBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert PEM to SM2 public key")
	}

	return parsedPubKey, nil
}

// Convert an `sm2.PublicKey` object to PEM formatted bytes.
func ConvertPublicKeyToPEM(pubKey *sm2.PublicKey) ([]byte, error) {
	pubKeyDer, err := x509.MarshalSm2PublicKey(pubKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert public key to PEM")
	}

	pubKeyPemBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyDer,
	}

	pubKeyPem := pem.EncodeToMemory(&pubKeyPemBlock)
	return pubKeyPem, nil
}

// Convert two big integers (a point on curve P256Sm2) to an `sm2.PublicKey` object.
func ConvertBigIntegersToPublicKey(x *big.Int, y *big.Int) (*sm2.PublicKey, error) {
	c := sm2.P256Sm2()
	if isOnCurve := c.IsOnCurve(x, y); !isOnCurve {
		return nil, fmt.Errorf("cannot convert big integers to public key because the point is not on curve P256Sm2")
	}

	pub := new(sm2.PublicKey)
	pub.Curve = c
	pub.X = x
	pub.Y = y

	return pub, nil
}
