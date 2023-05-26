# syntax=docker/dockerfile:1.4

FROM debian:bullseye-slim

RUN <<'EOF'
apt-get update
apt-get install -y --no-install-recommends \
  python3 \
  python3-requests \
  unzip \
  wget
rm -rf /var/lib/apt/lists/*
EOF

# install terraform.
# see https://www.terraform.io/downloads
ENV CHECKPOINT_DISABLE=1
RUN <<'EOF'
artifact_url=https://releases.hashicorp.com/terraform/1.4.6/terraform_1.4.6_linux_amd64.zip
artifact_path="/tmp/$(basename $artifact_url)"
wget -qO $artifact_path $artifact_url
unzip -o $artifact_path -d /usr/local/bin
rm $artifact_path
terraform version
EOF

VOLUME /host

WORKDIR /host

ENTRYPOINT ["./init/init.sh"]
