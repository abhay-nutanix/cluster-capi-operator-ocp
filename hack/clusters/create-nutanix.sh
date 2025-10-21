#!/bin/bash

function printcolor {
  COLOR='\033[0;32m'
  NC='\033[0m'
  printf "${COLOR}$1${NC}\n"
}

printcolor "Getting required variables"
export CLUSTER_NAME=$(kubectl get infrastructure cluster -o jsonpath="{.status.infrastructureName}")
export INFRASTRUCTURE_KIND=NutanixCluster
export CLUSTER_CONTROLPLANE_ENDPOINT=$(kubectl get infrastructure cluster -o jsonpath="{.status.apiServerInternalURI}" | sed -E "s|https?://(.*)$|\1|g")
export CLUSTER_CONTROLPLANE_HOST=$(echo ${CLUSTER_CONTROLPLANE_ENDPOINT} | cut -d':' -f1)
export CLUSTER_CONTROLPLANE_PORT=$(echo ${CLUSTER_CONTROLPLANE_ENDPOINT} | cut -d':' -f2)
export PRISM_CENTRAL_ADDRESS=$(kubectl get infrastructure cluster -o jsonpath="{.spec.platformSpec.nutanix.prismCentral.address}")
export PRISM_CENTRAL_PORT=$(kubectl get infrastructure cluster -o jsonpath="{.spec.platformSpec.nutanix.prismCentral.port}")

printcolor "Creating Nutanix infrastructure cluster"
envsubst <hack/clusters/templates/nutanix.yaml | kubectl apply -f -

printcolor "Creating core cluster"
envsubst <hack/clusters/templates/core.yaml | kubectl apply -f -

printcolor "Done"
