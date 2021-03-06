#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${SCRIPT_ROOT}/vendor/k8s.io/code-generator

if [ "$GOPATH" == "" ]; then
  echo "make sure GOPATH is set (case sensitive)"
  exit 1
fi

# dep ensure nukes the code-generator package.  just copy what we have here over to the vendor folder
mkdir -p $CODEGEN_PKG
cp -Rf code-generator/* $CODEGEN_PKG

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  kubernetes-grafana-controller/pkg/client kubernetes-grafana-controller/pkg/apis \
  grafana:v1alpha1 \
  --output-base "$(dirname ${BASH_SOURCE})/../../../src" \
  --go-header-file ${SCRIPT_ROOT}/codegen/boilerplate.go.txt

# To use your own boilerplate text use:
#   --go-header-file ${SCRIPT_ROOT}/codegen/custom-boilerplate.go.txt