# syntax=docker/dockerfile:1.4
# see https://github.com/wso2/product-is
# see https://github.com/wso2/product-is/releases/tag/v6.1.0
# see https://github.com/wso2/docker-is

# see https://github.com/wso2/product-is/releases
ARG WSO2IS_VERSION=6.1.0

# download the wso2is source code and build.
FROM maven:3.9.2-eclipse-temurin-11 as wso2is-build
ARG WSO2IS_VERSION
SHELL ["/bin/bash", "-euxo", "pipefail", "-c"]
WORKDIR /build
RUN <<EOF
apt-get update
apt-get install -y --no-install-recommends \
    unzip
rm -rf /var/lib/apt/lists/*
EOF
RUN <<EOF
wget -qO wso2is.tgz https://github.com/wso2/product-is/archive/refs/tags/v${WSO2IS_VERSION}.tar.gz
tar xf wso2is.tgz --strip-components=1
rm wso2is.tgz
EOF
RUN --mount=type=cache,target=/root/.m2 <<EOF
mvn clean install -DskipTests
EOF
RUN <<EOF
set -euxo pipefail
unzip modules/distribution/target/wso2is-${WSO2IS_VERSION}.zip
mkdir /wso2is
mv wso2is-${WSO2IS_VERSION}/* /wso2is
rmdir wso2is-${WSO2IS_VERSION}
EOF

# create image.
FROM eclipse-temurin:11
WORKDIR /wso2is
EXPOSE 9443
RUN <<EOF
apt-get update
apt-get install -y --no-install-recommends \
    unzip \
    vim
rm -rf /var/lib/apt/lists/*
EOF
COPY --from=wso2is-build /wso2is ./
ENTRYPOINT ["./bin/wso2server.sh"]
