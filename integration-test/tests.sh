#!/bin/bash
set -euo pipefail

trap clean_up ERR EXIT

function clean_up() {
    # Clean up Vault and Circle docker container
    docker rm -f vault circle > /dev/null 2>&1 || true

    # Clean up docker network
    docker network rm vaulttest > /dev/null 2>&1 || true
}

base_dir=$(dirname $0)

status_codes=(200 200 200 404 500)
grep_expressions=("circleci build is not currently running"
                  "provided VCS revision does not match the revision reported by circleci"
                  ""
                  '* 404: {"message":"Not Found","documentation_url":"https://developer.github.com/v3/repos/#get"}'
                  '* 500: An internal error occurred')

# Creating the Docker Network vaulttest
echo -n "Creating docker network: " ; docker network create vaulttest

# Creating the mock CircleCI server containers
for i in 1 2 3 4 5; do
    echo -n "Creating docker container for mock circleci server $i: "
    docker create --rm --name circle --network vaulttest \
            marcboudreau/dumb-server:latest \
            -sc ${status_codes[$((i-1))]} -resp /response
    docker cp $base_dir/responses/circle$i circle:/response
    echo -n "Starting docker container " ; docker start circle

    echo -n "Creating docker container for vault: "
    docker run \
            --rm \
            -d \
            --name vault \
            --network vaulttest \
            -e VAULT_TOKEN=root \
            -e VAULT_ADDR=http://127.0.0.1:8200 \
            -e VAULT_LOG_LEVEL=trace \
            vault-circleci-auth-plugin:test

    while ! docker exec vault vault auth list; do
        echo "Still waiting for Vault server to finish initializing..."
    done

    attempt_cache_expiry=
    if (( $i == 3 )); then
        attempt_cache_expiry="attempt_cache_expiry=5s"
    fi
    
    #sleep 1

    echo -n "Removing docker containers: "
    docker rm -f circle vault | tr '\n' ' '
done
