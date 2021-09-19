#!/bin/bash

helm --kube-context dev install redis -f ./values.yaml .
