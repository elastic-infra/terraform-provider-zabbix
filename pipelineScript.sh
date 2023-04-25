#!/bin/bash

set -e
set +x

function provider() {
  echo -n $(jq -r .namespace ../version.json)/$(jq -r .provider ../version.json)
}

function version_le() {
  [[ "$(echo -e "$1\n$2"  | sort -Vr | head -n 1)" == "${2}" ]]
}

function check() {
  version_le "$(jq -r .version ../version.json)" "$(curl -sL ${CITIZEN_ADDR}/v1/providers/$(provider)/versions  | jq -r ${JQ}  | sort -Vr | head -n1)"
}

function populateToCitizen() {
  /usr/local/bin/citizen provider \
    $(jq -r .namespace ../version.json) \
    $(jq -r .provider ../version.json) \
    $(jq -r .version ../version.json) \
    4.1,5.0
}

echo "nameserver 1.1.1.1" > /etc/resolv.conf
pushd /workspace/


if check
then
  echo "|_ [-] No need to push a new provider $(provider)"
else
  populateToCitizen
  echo "|_ [+] Populated $(provider)"
fi
echo
popd


echo -e "\n[+] Done ! bye\n"