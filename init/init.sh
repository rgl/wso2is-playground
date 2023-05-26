#!/bin/bash
set -euxo pipefail

# initilize wso2is.
install -d tmp
python3 \
    init/init.py \
    wso2is-init \
    --base-url https://wso2is.test:9443

# initialize keycloak.
export CHECKPOINT_DISABLE=1
export TF_LOG=DEBUG
export TF_LOG_PATH=/host/terraform.log
pushd init/keycloak
terraform init
terraform apply -auto-approve
popd
