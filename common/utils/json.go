package utils

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var jsonpbMarshaler = &protojson.MarshalOptions{
	UseProtoNames:   true,
	EmitUnpopulated: true,
}

func JsonToString(v interface{}) string {
	if v, ok := v.(proto.Message); ok {
		s, err := jsonpbMarshaler.Marshal(v)
		if err != nil {
			return ""
		}
		return string(s)
	}

	ret, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(ret)
}

func JsonToByte(v interface{}) []byte {
	if v, ok := v.(proto.Message); ok {
		s, err := jsonpbMarshaler.Marshal(v)
		if err != nil {
			return nil
		}
		return s
	}

	ret, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return ret
}
