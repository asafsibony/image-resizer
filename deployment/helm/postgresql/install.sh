#!/bin/bash

helm --kube-context dev install postgresql -f ./values.yaml .
