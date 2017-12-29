package hoji

import (
	"bytes"
	"encoding/binary"
	"log"

	"gitlab.com/rodzzlessa24/hoji/base58"
)

// IntToByte converts an int64 to a byte array
func IntToByte(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

//ExtractPubKey is
func ExtractPubKey(address []byte) []byte {
	decodeAddr := base58.Decode(address)
	return decodeAddr[1 : len(decodeAddr)-4]
}
