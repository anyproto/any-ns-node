SHELL=/bin/bash
export GOPRIVATE=github.com/anyproto
export PATH:=deps:$(PATH)
export CGO_ENABLED:=1
BUILD_GOOS:=$(shell go env GOOS)
BUILD_GOARCH:=$(shell go env GOARCH)

ifeq ($(CGO_ENABLED), 0)
	TAGS:=-tags nographviz
else
	TAGS:=
endif

.PHONY: prereqs-for-mac
prereqs-for-mac:
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	brew install protobuf protoc-gen-go protoc-gen-go-grpc
	brew install clang-format
	brew install grpcurl
	brew install golangci-lint

.PHONY: deps
deps:
	go install go.uber.org/mock/mockgen@latest
	go mod download
	go build -o deps/protoc-gen-go-drpc storj.io/drpc/cmd/protoc-gen-go-drpc
	go build -o deps/protoc-gen-gogofaster github.com/gogo/protobuf/protoc-gen-gogofaster
	go build -o deps github.com/ahmetb/govvv

.PHONY: build
build:
	@$(eval FLAGS := $$(shell PATH=$(PATH) govvv -flags -pkg github.com/anyproto/any-sync/app))
	GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build $(TAGS) -v -o bin/any-ns-node -ldflags "$(FLAGS) -X github.com/anyproto/any-sync/app.AppName=any-ns-node" github.com/anyproto/any-ns-node/cmd

contracts/mock/contracts_mock.go: contracts/contracts.go
	# go install go.uber.org/mock/mockgen@latest
	mockgen -source=contracts/contracts.go > contracts/mock/contracts_mock.go

account_abstraction/mock/account_abstraction_mock.go: account_abstraction/account_abstraction.go
	mockgen -source=account_abstraction/account_abstraction.go > account_abstraction/mock/account_abstraction_mock.go

alchemysdk/mock/alchemysdk_mock.go: alchemysdk/alchemysdk.go
	mockgen -source=alchemysdk/alchemysdk.go > alchemysdk/mock/alchemysdk_mock.go

cache/mock/cache_mock.go: cache/cache.go
	mockgen -source=cache/cache.go > cache/mock/cache_mock.go

nonce_manager/mock/nonce_manager_mock.go: nonce_manager/nonce_manager.go
	mockgen -source=nonce_manager/nonce_manager.go > nonce_manager/mock/nonce_manager.go

queue/mock/queue_mock.go: queue/queue.go
	mockgen -source=queue/queue.go > queue/mock/queue_mock.go

db/mock/db_mock.go: db/db.go
	mockgen -source=db/db.go > db/mock/db_mock.go

.PHONY: mocks
mocks: contracts/mock/contracts_mock.go account_abstraction/mock/account_abstraction_mock.go alchemysdk/mock/alchemysdk_mock.go cache/mock/cache_mock.go nonce_manager/mock/nonce_manager_mock.go queue/mock/queue_mock.go db/mock/db_mock.go

.PHONY: test
test: mocks
	go test ./... --cover $(TAGS)

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

# TODO: fix all ../deployments directories!
# TODO: remove "sepolia" network dependency
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

anytype_crypto/anytype_registrar_stub.go: ../deployments/sepolia/AnytypeRegistrarImplementation.json
	node utils/combined_json_to_abi_and_bin.js ../deployments/sepolia/AnytypeRegistrarImplementation.json
	abigen --bin contract.bin --abi contract.abi --pkg anytype_crypto --type AnytypeRegistrarImplementation --out anytype_crypto/anytype_registrar_stub.go
	rm contract.bin contract.abi

anytype_crypto/scw_stub.go: ./abis/SCW_ABI.json
	abigen --abi contract.abi --pkg anytype_crypto --type SCW --out anytype_crypto/scw_stub.go

# Go stubs generated from *.sol files
anytype_crypto:\
    anytype_crypto/ens_registry_stub.go\
    anytype_crypto/name_wrapper_stub.go\
    anytype_crypto/anytype_resolver_stub.go\
    anytype_crypto/anytype_controller_private_stub.go\
    anytype_crypto/anytype_registrar_stub.go\
    anytype_crypto/scw_stub.go

# Build everything
.PHONY: all
all: deps anytype_crypto build

# Run a dRPC server
.PHONY: run
run:
	go run ./cmd --c=config.yml

# Run a test client that connects to server
.PHONY: is-name-avail
is-name-avail:
	grpcurl -plaintext -d '{"fullName": "hello.any"}' localhost:8080 Anyns/IsNameAvailable

.PHONY: reg-test-name
reg-test-name:
	grpcurl -plaintext -d '{"fullName": "some10.any", "ownerEthAddress": "0xe595e2BA3f0cE990d8037e07250c5C78ce40f8fF", "ownerAnyAddress": "A6WVkd1MxX1i7hGQCcDhMFvfEzokPppRzxve2wdhTZ8jZTio"}' localhost:8080 Anyns/NameRegister

.PHONY: aa-get-user-account
aa-get-user-account:
	grpcurl -plaintext -d '{"ownerEthAddress": "0xE34230c1f916e9d628D5F9863Eb3F5667D8FcB37"}' localhost:8080 AnynsAccountAbstraction/GetUserAccount

