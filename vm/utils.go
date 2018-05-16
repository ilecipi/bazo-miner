package vm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"
)

const UINT16_MAX uint16 = 65535

func UInt64ToByteArray(element uint64) []byte {
	ba := make([]byte, 8)
	binary.LittleEndian.PutUint64(ba, uint64(element))
	return ba
}

func UInt16ToByteArray(element uint16) []byte {
	ba := make([]byte, 2)
	binary.LittleEndian.PutUint16(ba, uint16(element))
	return ba
}

func ByteArrayToUI16(element []byte) (uint16, error) {
	if bytes.Equal([]byte{}, element) {
		return 0, nil
	}
	return binary.LittleEndian.Uint16(element), nil
}

func StrToBigInt(element string) big.Int {
	var result big.Int
	hexEncoded := hex.EncodeToString([]byte(element))
	result.SetString(hexEncoded, 16)
	return result
}

func ByteArrayToInt(element []byte) int {
	ba := make([]byte, 64-len(element))
	ba = append(element, ba...)
	return int(binary.LittleEndian.Uint64(ba))
}

func BigIntToString(element big.Int) string {
	ba := element.Bytes()
	return string(ba[:])
}
