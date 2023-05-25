#!/bin/bash
set -euxo pipefail

# initilize wso2is.
install -d tmp
python3 \
    init/init.py \
    wso2is-init \
    --base-url https://wso2is.test:9443
