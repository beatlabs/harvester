#!/usr/bin/env bash

#if [ -z "${CODECOV_TOKEN}" ]; then
#  echo "CODECOV_TOKEN is not set, skipping code coverage upload.."
#else
  echo "Running code coverage upload:"
  curl -s https://codecov.io/bash > .codecov && chmod +x .codecov
  ./.codecov -X fix
  echo "Done."
#fi
