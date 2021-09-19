#!/bin/sh

helm install ingress-nginx --kube-context dev -f values.yaml .
