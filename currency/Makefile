.PHONY: protos

protos:
	protoc -I protos/ protos/currency.proto --go_out=protos --go-grpc_out=require_unimplemented_servers=false:protos