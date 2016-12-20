#!/bin/bash -eu
echo "----------" 
echo "$@" 
echo "----------" 

ginkgo -randomizeSuites=true -randomizeAllSpecs=true -keepGoing=true -r "$@"
