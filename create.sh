#!/bin/bash
set -euo pipefail

# destroy the existing environment.
docker compose down --remove-orphans --volumes
rm -f terraform.{log,tfstate,tfstate.backup} tfplan

# build the environment.
docker compose --profile client build --progress plain
docker compose --profile test build --progress plain
docker compose build --progress plain

# start the environment in background.
docker compose up --detach

# wait for the services to exit.
function wait-for-service {
  echo "Waiting for the $1 service to complete..."
  while true; do
    result="$(docker compose ps --all --status exited --format json $1)"
    if [ -n "$result" ] && [ "$result" != 'null' ]; then
      exit_code="$(jq -r '.[].ExitCode' <<<"$result")"
      break
    fi
    sleep 3
  done
  docker compose logs $1
  return $exit_code
}
wait-for-service init

# start the client services.
docker compose --profile client up --detach

# execute the automatic tests.
cat <<'EOF'

#### Automated tests results

EOF
echo 'example-go-confidential client test:'
docker compose run example-go-confidential-test | sed -E 's,^(.*),  \1,g'
echo

# show how to use the system.
cat <<'EOF'

#### Manual tests

example-go-confidential client:
  Start the login dance at http://example-go-confidential.test:8081 as admin:admin

wso2is:
  https://wso2is.test:9443/console/ (WSO2IS Console; login as `admin`:`admin`)
  https://wso2is.test:9443/carbon/ (WSO2IS Carbon Management Console; login as `admin`:`admin`)
  https://wso2is.test:9443/myaccount/ (WSO2IS My Account; login as `admin`:`admin`)
  https://wso2is.test:9443/oauth2/token/.well-known/openid-configuration
EOF
