package gen

//go:generate protoc -I ../../../_proto --go_out=../../../ --go_opt=module=github.com/Elessarov1/geocoder-go --go-grpc_out=../../../ --go-grpc_opt=module=github.com/Elessarov1/geocoder-go ../../../_proto/geocoder/v1/geocoder.proto
