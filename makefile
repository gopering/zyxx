.PHONY: all

NAME=zyxx
BASE_DIR=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
RELEASE=${BASE_DIR}/release
TARGET_BIN=${RELEASE}/bin
TARGET_CONF=${RELEASE}/conf

DATE=$(shell date +"%Y%m%d%H%M%S")

GOBUILD= CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

define compile_and_pack
    @rm -rf ${RELEASE}
	@mkdir -p ${TARGET_BIN}
	@mkdir -p ${TARGET_CONF}
	$(GOBUILD) -o ${TARGET_BIN}/$(NAME) -ldflags "${FLAGS}" ${BASE_DIR}/main.go
	cp -rf ${BASE_DIR}/etc/* ${TARGET_CONF}
endef

all:
	$(call compile_and_pack)

clean:
	@rm -rf ${RELEASE}

fmt:
	gofmt -s -w .

lint:
	golangci-lint run
