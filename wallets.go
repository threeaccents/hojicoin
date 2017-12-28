package hoji

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

//Wallets is
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	if err := wallets.LoadFromFile(); err != nil {
		return nil, err
	}

	return &wallets, nil
}

// AddWallet adds a Wallet to Wallets
func (ws *Wallets) AddWallet() ([]byte, error) {
	wallet, err := NewWallet()
	if err != nil {
		return nil, err
	}
	address, err := wallet.GetAddress()
	if err != nil {
		return nil, err
	}

	ws.Wallets[fmt.Sprintf("%s", address)] = wallet

	return address, nil
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws *Wallets) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	if len(fileContent) == 0 {
		return nil
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	if err := gob.NewDecoder(bytes.NewReader(fileContent)).Decode(&wallets); err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() error {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(walletFile, content.Bytes(), 0644)
}
