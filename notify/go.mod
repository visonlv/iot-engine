module github.com/visonlv/iot-engine/notify

go 1.20

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/envoyproxy/protoc-gen-validate v1.0.2
	// github.com/visonlv/go-vkit v0.0.0-incompatible
	github.com/visonlv/go-vkit v0.0.0-20231019071952-551a043011db
	google.golang.org/genproto/googleapis/api v0.0.0-20231016165738-49dd2c1f3d0b
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	gorm.io/gorm v1.25.5 // indirect
)

require github.com/nats-io/nats.go v1.29.0

require (
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/nats-io/jwt/v2 v2.5.3 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20231012201019-e917dd12ba7a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231012201019-e917dd12ba7a // indirect
	gorm.io/driver/mysql v1.5.1 // indirect
)

// replace github.com/visonlv/go-vkit v0.0.0-incompatible => ../../../visonlv/go-vkit
