package utils

import (
	"encoding/binary"
	"encoding/hex"

	"github.com/visonlv/go-vkit/utilsx"
)

var chars = []string{"a", "b", "c", "d", "e", "f",
	"g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s",
	"t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5",
	"6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "G", "H", "I",
	"J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V",
	"W", "X", "Y", "Z"}

func GenShortId() string {
	var appid string
	uuid := utilsx.GenUuid()
	for i := 0; i < 8; i++ {
		str := uuid[i*4 : 4*(i+1)]
		hexStr, _ := hex.DecodeString(str)
		num := binary.BigEndian.Uint16(hexStr)
		appid += chars[num%0x3E]
	}
	return appid
}
