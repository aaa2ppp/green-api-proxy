TMP_DIR := ./tmp

BUILD_OPTIONS ?=

MERGE_FILES ?= Makefile *.go *.sh *.md *.example
SRC ?= .
DST ?= 1

.PHONY: all deps build run cert merge patch clean

all: deps build

deps:
	go mod tidy

build:
	go build $(BUILD_OPTIONS) -o ./bin/proxy ./cmd/proxy

run:
	./bin/proxy

cert:
	./scripts/gencert.sh


MERGE_FIND_PARTS := $(patsubst %,-o -name '%',$(MERGE_FILES))
MERGE_FIND_EXPR := $(wordlist 2,$(words $(MERGE_FIND_PARTS)),$(MERGE_FIND_PARTS))

merge:
	@mkdir -p $(TMP_DIR)
	./scripts/merge-code.sh $(SRC) > $(TMP_DIR)/$(DST).code
	

# Создает прекоммит патч
patch:
	@mkdir -p $(TMP_DIR)
	git diff --staged -- $(SRC) > $(TMP_DIR)/$(DST).patch


clean:
	-rm -fr ./bin ./cert $(TMP_DIR)
