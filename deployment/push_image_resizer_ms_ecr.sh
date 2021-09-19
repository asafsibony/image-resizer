#!/bin/bash

docker build -t 353322593157.dkr.ecr.us-west-2.amazonaws.com/image-resizer/image-resizer-ms:v1 -f cmd/image-resizer-ms/Dockerfile .
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 353322593157.dkr.ecr.us-west-2.amazonaws.com
docker push 353322593157.dkr.ecr.us-west-2.amazonaws.com/image-resizer/image-resizer-ms:v1