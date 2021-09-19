#!/bin/bash

helm --kube-context dev install kafka -f ./values.yaml .
