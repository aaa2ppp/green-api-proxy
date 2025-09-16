PORT ?= 8080
TLS  ?= off
CN   ?= localhost
IP   ?= 

LOG_LEVEL    ?= INFO
ALLOW_ORIGIN ?= 

BUILD_OPTS ?= -ldflags "-s -w" -trimpath
RUN_OPTS   ?= --restart=unless-stopped --cpus=0.1 --memory=50m

GENCERT_CMD := $(PWD)/../scripts/gencert.sh

all: up

.build.done:
	-rm -fr ./bin
	GOOS=linux GOARCH=amd64 go build $(BUILD_OPTIONS) -o ./bin/app ../cmd/proxy
	#upx --lzma --force-overwrite -q --no-progress ./bin/app
	@touch $@

.cert.done:
	-rm -fr cert
	mkdir -p cert/cert && chmod 0700 cert
	cd cert && CN=$(CN) IP=$(IP) $(GENCERT_CMD) && chmod 0755 ./cert && chmod 0644 ./cert/*
	@touch $@

.build-image.done:
	docker build -t green-api-proxy:latest .
	@touch $@

build: .build.done

cert: .cert.done

build-image: .build-image.done

up: cert build build-image
	docker run -d $(RUN_OPTIONS) -p $(PORT):$(PORT) \
		-e SERVER_ADDR=:$(PORT) \
		-e SERVER_TLS=$(TLS) \
		-e ALLOW_ORIGIN=$(ALLOW_ORIGIN) \
		-e LOG_LEVEL=$(LOG_LEVEL) \
		--name green-api-proxy green-api-proxy

down: stop
	-docker container rm green-api-proxy

clean: down
	-docker image rm green-api-proxy
	-rm -fr ./bin ./cert ./.*.done

start:
	docker start green-api-proxy
	
stop:
	-docker stop green-api-proxy
