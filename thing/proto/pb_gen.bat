@echo on

protoc -I=. -I=../../common/proto/ --go_out=. --vkit_out=. --openapiv2_out=. --openapiv2_opt=disable_default_errors=true,json_names_for_fields=false --vkit_opt=--handlePath=../handler --validate_out="lang=go:."  thing.proto
copy thing.swagger.json ..\\..\\tools\\swagger-apihub\\swaggerui\\config\\

exit