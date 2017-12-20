package hoji

import (
	"bytes"
	"encoding/binary"
	"log"
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
