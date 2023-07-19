export GOPRIVATE=github.com/anyproto
export PATH:=deps:$(PATH)

########### Vars:
PORT=8080
GETH_URL="https://sepolia.infura.io/v3/68c55936b8534264801fa4bc313ff26f" 
# TODO: read from ../deployments/sepolia folder
CONTRACT_REG_ADDR="0xc0D3c96aE923Da6b45E6d4c21a0424730a20BCA9" 
CONTRACT_NW_ADDR="0xFe69BF9B3fD69d09977b37b5953C8B43687f3B23"
CONTRACT_RESOLVER_ADDR="0x34F9c5CB9b6dcc036e045a15af20CEdC0dE4dcB2"
CONTRACT_CONTROLLER_PRIVATE_ADDR="0x45bA047AD44e35FbF5A1375F79ea3872ceDB1732"
CONTRACT_RESOLVER_ADDR="0x34F9c5CB9b6dcc036e045a15af20CEdC0dE4dcB2"

#.PHONY: build
#build:
#	@$(eval FLAGS := $$(shell PATH=$(PATH) govvv -flags -pkg github.com/anyproto/any-sync/app))
#	go build -v -o bin/anyns-node -ldflags "$(FLAGS) -X github.com/anyproto/anyns-node/app.AppName=anyns-node" github.com/anyproto/anyns-node/cmd

.PHONY: test
test:
	go test ./... --cover

.PHONY: deps
deps:
	go mod download
	go build -o deps/protoc-gen-go-drpc storj.io/drpc/cmd/protoc-gen-go-drpc
	go build -o deps/protoc-gen-gogofaster github.com/gogo/protobuf/protoc-gen-gogofaster
	go build -o deps github.com/ahmetb/govvv


.PHONY: prereqs-for-mac
prereqs-for-mac:
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	brew install protobuf protoc-gen-go protoc-gen-go-grpc
	brew install clang-format
	brew install grpcurl
	brew install golangci-lint

# Generate protobuf stubs and documentation
pb: proto/anyns_api_server.proto
	#protoc --proto_path=proto proto/*.proto --gofast_out=. --go-grpc_out=.
	protoc --gogofaster_out=:. --go-drpc_out=protolib=github.com/gogo/protobuf:. proto/*.proto
	protoc --doc_out=./doc --doc_opt=html,index.html proto/*.proto

.PHONY: check-style
check-style:
	golangci-lint run -E errcheck -E gofmt -E revive 

# This doesn work unfortunately:
#abigen --combined-json contracts/build/contracts/XXX.json --pkg bootnodes --out xxx.go
#abigen --sol contracts/contracts/XXX.sol --pkg bootnodes --out xxx.go
anytype_crypto/ens_registry_stub.go: ../deployments/sepolia/ENSRegistry.json
	# convert combined json -> ./contract.abi and ./contract.bin
	node utils/combined_json_to_abi_and_bin.js ../deployments/sepolia/ENSRegistry.json
	# now generate Golang stub
	abigen --bin contract.bin --abi contract.abi --pkg anytype_crypto --type ENSRegistry --out anytype_crypto/ens_registry_stub.go
	rm contract.bin contract.abi

anytype_crypto/name_wrapper_stub.go: ../deployments/sepolia/AnytypeNameWrapper.json
	node utils/combined_json_to_abi_and_bin.js ../deployments/sepolia/AnytypeNameWrapper.json
	abigen --bin contract.bin --abi contract.abi --pkg anytype_crypto --type AnytypeNameWrapper --out anytype_crypto/name_wrapper_stub.go
	rm contract.bin contract.abi

anytype_crypto/anytype_resolver_stub.go: ../deployments/sepolia/AnytypeResolver.json
	node utils/combined_json_to_abi_and_bin.js ../deployments/sepolia/AnytypeResolver.json
	abigen --bin contract.bin --abi contract.abi --pkg anytype_crypto --type AnytypeResolver --out anytype_crypto/anytype_resolver_stub.go
	rm contract.bin contract.abi

anytype_crypto/anytype_controller_private_stub.go: ../deployments/sepolia/AnytypeRegistrarControllerPrivate.json
	node utils/combined_json_to_abi_and_bin.js ../deployments/sepolia/AnytypeRegistrarControllerPrivate.json
	abigen --bin contract.bin --abi contract.abi --pkg anytype_crypto --type AnytypeRegistrarControllerPrivate --out anytype_crypto/anytype_controller_private_stub.go
	rm contract.bin contract.abi

# requires Go stubs generated from *.sol files
anytype_crypto:\
    anytype_crypto/ens_registry_stub.go\
    anytype_crypto/name_wrapper_stub.go\
    anytype_crypto/anytype_resolver_stub.go\
    anytype_crypto/anytype_controller_private_stub.go

# Build everything
.PHONY: all
all: deps pb anytype_crypto

# Run a gRPC server
.PHONY: run-server
run-server: anytype_crypto
	GRPC_PORT=$(PORT)\
    GETH_URL=$(GETH_URL)\
    CONTRACT_REG_ADDR=$(CONTRACT_REG_ADDR)\
    CONTRACT_NW_ADDR=$(CONTRACT_NW_ADDR)\
    CONTRACT_RESOLVER_ADDR=$(CONTRACT_RESOLVER_ADDR)\
    CONTRACT_CONTROLLER_PRIVATE_ADDR=$(CONTRACT_CONTROLLER_PRIVATE_ADDR)\
    CONTRACT_RESOLVER_ADDR=$(CONTRACT_RESOLVER_ADDR)\
    go run .

# Run a test client that connects to server
.PHONY: is-name-avail
is-name-avail:
	grpcurl -plaintext -d '{"fullName": "hello.any"}' localhost:8080 Anyns/IsNameAvailable

.PHONY: reg-test-name
reg-test-name:
	grpcurl -plaintext -d '{"fullName": "some10.any", "ownerEthAddress": "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF", "ownerAnyAddress": "A6WVkd1MxX1i7hGQCcDhMFvfEzokPppRzxve2wdhTZ8jZTio"}' localhost:8080 Anyns/NameRegister

# Print a helpful description of the gRPC interface
.PHONY: describe
describe:
	grpcurl -plaintext localhost:$(PORT) describe Anyns

