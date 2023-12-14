# iot-engine

## 版本表

| 软件               | 使用版本 |
| ------------------ | -------- |
| GO                 | 1.20.8   |
| EMQX               | 5.3.0    |
| NATS               | 2.10.2   |
| protoc             | 3.20.1   |
| protoc-gen-go      | 1.28     |
| protoc-gen-go-grpc | 1.2      |

## go开发环境

### 安装参数校验工具
```
go install github.com/envoyproxy/protoc-gen-validate@v1.0.2

```


### 安装API文档生成工具
```
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-openapiv2@v2.18.0

```

### 安装vkit grpc接口生成工具
```
go install github.com/visonlv/protoc-gen-vkit@master

```

### 安装protoc
```
https://github.com/protocolbuffers/protobuf/releases/download/v3.20.3/protoc-3.20.3-win64.zip

```

### 安装protoc go
```
https://pkg.go.dev/google.golang.org/protobuf@v1.28.0
https://github.com/protocolbuffers/protobuf-go/releases/tag/v1.28.0
```

### 文档导览
* [物模型基本结构](doc/model.md)* 
* [影子基本机构](doc/shadow.md)* 
* [设备主题定义](doc/设备主题定义.md)* 
* [系统主题定义](doc/系统主题定义.md)* 