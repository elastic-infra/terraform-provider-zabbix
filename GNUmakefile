TEST?="./zabbix"
PKG_NAME=zabbix
DIR=~/.terraform.d/plugins

default: build


build:
	go install

install:
	mkdir -vp $(DIR)
	go build -o $(DIR)/terraform-provider-zabbix

uninstall:
	@rm -vf $(DIR)/terraform-provider-zabbix


test:
	go test $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m
