#!/bin/sh

nohup kubectl port-forward consul-server-0 8501:8500 > /dev/null 2>&1 &