.PHONY: protos

protos:
		protoc --proto_path=. ./*.proto --go_out=../
		protoc --proto_path=. ./*.proto --go-grpc_out=../