BUILD_OPTIONS ?=

all: build

build:
	go build $(BUILD_OPTIONS) -o ./bin/app .

run:
	./bin/app

.PHONY: cert

cert:
	./scripts/gencert.sh
