package utils

import (
	"hash/fnv"

	"github.com/visonlv/iot-engine/common/define"
)

func GetGroupId(s string) int32 {
	return GetHashSlotId(s, int32(define.MaxGroup+1))
}

func GetHashSlotId(s string, max int32) int32 {
	maxGroup := uint32(max)
	h := fnv.New32a()
	h.Write([]byte(s))
	hash := h.Sum32()
	slot := hash % maxGroup
	return int32(slot)
}
