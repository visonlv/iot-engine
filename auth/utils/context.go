package utils

import (
	"bytes"
	"context"
	"runtime"
	"strconv"

	"google.golang.org/grpc/metadata"
)

func GetUserIdFromContext(ctx context.Context) string {
	return GetValueFromContext(ctx, "user_id")
}

func GetRoleCodeFromContext(ctx context.Context) string {
	return GetValueFromContext(ctx, "role_code")
}

func GetValueFromContext(ctx context.Context, key string) string {
	gmd, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		gmd = metadata.MD{}
	}
	values := gmd.Get(key)
	if values == nil || len(values) <= 0 {
		return ""
	}
	return values[0]
}

func GetGid() (gid uint64) {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		panic(err)
	}
	return n
}
