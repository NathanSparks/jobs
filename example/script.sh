#!/bin/bash

INPUTDIR=${1}
RUNNO=${2}

source master.sh

echo pwd = $PWD
echo
echo LD_LIBRARY_PATH = $LD_LIBRARY_PATH
echo
echo PATH = $PATH
echo

cp -R ${INPUTDIR} .
mv ${RUNNO} data
root -b -q run.C
