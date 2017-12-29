package hoji

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"gitlab.com/rodzzlessa24/hoji/base58"
	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

//Wallet is
type Wallet struct {
	PublicKey  []byte
	PrivateKey *ecdsa.PrivateKey
}

//NewWallet creates a new private public key pair
func NewWallet() (*Wallet, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	wallet := Wallet{
		PrivateKey: private,
		PublicKey:  pubKey,
	}

	return &wallet, nil
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	pubKeyHash := base58.Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

//GetAddress is
func (w *Wallet) GetAddress() ([]byte, error) {
	hashedPubKey, err := hashPubKey(w.PublicKey)
	if err != nil {
		return nil, err
	}
	versionedPayload := append([]byte{version}, hashedPubKey...)
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)

	address := base58.Encode(fullPayload)

	return address, nil
}

func hashPubKey(pubKey []byte) ([]byte, error) {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		return nil, err
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160, nil
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}
