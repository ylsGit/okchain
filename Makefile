maDEP := $(shell command -v dep 2> /dev/null)
SUM := $(shell which shasum)

COMMIT := $(shell git rev-parse HEAD)
CAT := $(if $(filter $(OS),Windows_NT),type,cat)

Version=v0.10.0
CosmosSDK=v0.37.8
Tendermint=v0.32.9
Iavl=v0.12.4
ServerName=okchaind
ClientName=okchaincli

# process linker flags
ifeq ($(VERSION),)
    VERSION = $(COMMIT)
endif

build_tags = netgo

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

ldflags = -X github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.Version=$(Version) \
  -X github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.ServerName=$(ServerName) \
  -X github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.ClientName=$(ClientName) \
  -X github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
  -X github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.CosmosSDK=$(CosmosSDK) \
  -X github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.Tendermint=$(Tendermint) \
  -X github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.VendorDirHash=$(shell $(SUM) -a 256 go.sum | cut -d ' ' -f1) \
  -X "github.com/okex/okchain/vendor/github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags)"


ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -ldflags '$(ldflags)'


all: install

install: okchain

okchain:
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/okchaind
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/okchaincli

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

get_vendor_deps:
	@echo "--> Generating vendor directory via dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -vendor-only

update_vendor_deps:
	@echo "--> Running dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -update


cli:
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/okchaincli

server:
	go install -v $(BUILD_FLAGS) -tags "$(BUILD_TAGS)" ./cmd/okchaind

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s

build:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/okchaind.exe ./cmd/okchaind
	go build $(BUILD_FLAGS) -o build/okchaincli.exe ./cmd/okchaincli
else
	go build $(BUILD_FLAGS) -o build/okchaind ./cmd/okchaind
	go build $(BUILD_FLAGS) -o build/okchaincli ./cmd/okchaincli
endif

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-docker-okchainnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/okchaind/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/okchaind:Z okchain/node testnet --v 4 -o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down


.PHONY: build
