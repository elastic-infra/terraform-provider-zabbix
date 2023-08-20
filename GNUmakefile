TEST?="./zabbix"
PKG_NAME=zabbix
#CDIR=citizen/

VERSION=1.0.0

HOSTNAME=terraform.local
NAMESPACE=local
NAME=zabbix
BINARY=terraform-provider-${NAME}
DIR=~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}
OS_ARCHs := darwin_arm64 linux_amd64


default: build


build:
	go build -o ${BINARY}

install: build
	@$(foreach	OS_ARCH,$(OS_ARCHs), \
        echo Processing $(OS_ARCH); \
        mkdir -p ${DIR}/${VERSION}/${OS_ARCH}; \
		cp ${BINARY} ${DIR}/${VERSION}/${OS_ARCH}; \
    ) \
    rm terraform-provider-zabbix

uninstall:
	@rm -rf ${DIR}


test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

release:
	goreleaser release --clean

#citizen:
#	apt update
#	apt install -y zip jq
#	go get github.com/atypon/go-zabbix-api
#	go get github.com/atypon/terraform-provider-zabbix/zabbix
#	go build -o terraform-provider-zabbix
#	zip -r gcp-zabbix_`jq -r .version version.json`_linux_amd64.zip  \
#		terraform-provider-zabbix
##	shasum -a 256 $(CDIR)*.zip > $(CDIR)gcp-zabbix_`jq -r .version version.json`_SHA256SUMS
##	gpg --batch --gen-key gen-key-script
##	gpg --detach-sign $(CDIR)gcp-zabbix_`jq -r .version version.json`_SHA256SUMS
