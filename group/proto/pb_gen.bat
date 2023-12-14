@echo on

protoc -I=. -I=../../common/proto/ --go_out=. --vkit_out=. --openapiv2_out=. --vkit_opt=--handlePath=../handler --validate_out="lang=go:."  group.proto
copy group.swagger.json ..\\..\\tools\\swagger-apihub\\swaggerui\\config\\



exit