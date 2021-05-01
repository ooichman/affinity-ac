#!/bin/bash

if [[ -z $1 ]]; then

	echo "No Package TAG provided"
else
	buildah bud -f Dockerfile -t $1 && \
	buildah push $1
fi