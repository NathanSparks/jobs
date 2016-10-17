#!/bin/bash
ENVIRONMENT=${1}
INDIR=${2}
RUNNO=${3}

source $ENVIRONMENT

echo pwd = $PWD
echo
echo LD_LIBRARY_PATH = $LD_LIBRARY_PATH
echo
echo PATH = $PATH
echo

cp -R ${INDIR} .
mv ${RUNNO} data
root -b -q run.C
