TEST?="./zabbix"
PKG_NAME=zabbix
DIR=~/.terraform.d/plugins/terraform.local/local/zabbix/1.0.0/linux_amd64/
CDIR=citizen/

default: build


build:
	go install

install:
	mkdir -vp $(DIR)
	go build -o $(DIR)/terraform-provider-zabbix_v1.0.0
	chmod +x $(DIR)/terraform-provider-zabbix_v1.0.0

uninstall:
	@rm -vf $(DIR)/terraform-provider-zabbix_v1.0.0


test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

citizen:
	apt update
	apt install -y zip jq
	rm -rf $(CDIR)
	mkdir -vp $(CDIR)
	go build -o $(CDIR)terraform-provider-zabbix_`jq -r .version version.json`
	zip -r $(CDIR)gcp-zabbix_`jq -r .version version.json`_linux_amd64.zip  \
		$(CDIR)terraform-provider-zabbix_`jq -r .version version.json`
#	shasum -a 256 $(CDIR)*.zip > $(CDIR)gcp-zabbix_`jq -r .version version.json`_SHA256SUMS
#	gpg --batch --gen-key gen-key-script
#	gpg --detach-sign $(CDIR)gcp-zabbix_`jq -r .version version.json`_SHA256SUMS