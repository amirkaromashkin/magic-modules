#!/usr/bin/env bash

if [ -z "${TGC_PATH}" ]; then
    echo "set TGC_PATH to absolute path of TGC repo";
    exit 1;
fi

rm -rf  $TGC_PATH/cai2hcl/*

bundle exec compiler.rb \
    -e terraform -o $TGC_PATH/cai2hcl \
    -v beta \
    -p products/compute \
    -t ForwardingRule,GlobalForwardingRule,RegionBackendService,BackendService,HealthCheck,RegionHealthCheck \
    -f tgc_cai2hcl


cd $TGC_PATH

go test ./cai2hcl/...