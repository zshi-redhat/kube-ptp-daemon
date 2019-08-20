#!/bin/bash

vendor/k8s.io/code-generator/generate-groups.sh all \
github.com/zshi-redhat/kube-ptp-daemon/pkg/client \ github.com/zshi-redhat/kube-ptp-daemon/pkg/apis \
ptp:v1
