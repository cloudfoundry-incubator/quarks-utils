#!/bin/bash

get_quarks_utils_bin() {
  GIT_ROOT=${GIT_ROOT:-$(git rev-parse --show-toplevel)}
  cd $GIT_ROOT

  # Gets the quarks-utils version from go.mod
  quarks_utils_version=$(grep code.cloudfoundry.org/quarks-utils go.mod | awk '{print $2}' | head -1)
  quarks_utils_version=$(echo $quarks_utils_version | sed 's/.*-//g') # Keep only the commit
  if [ "$quarks_utils_version" = "utils" ]; then
    echo "failed to parse commit from go.mod"
    exit 1
  fi

  if [ -d /usr/local/opt/gnu-tar/libexec/gnubin ]; then
    PATH="/usr/local/opt/gnu-tar/libexec/gnubin:$PATH"
  fi

  mkdir -p tools/quarks-utils
  wget -qO - https://github.com/cloudfoundry-incubator/quarks-utils/archive/"$quarks_utils_version".tar.gz | \
    tar -C tools/quarks-utils -xz --strip-components=1 --no-anchored bin/
}

get_quarks_utils_bin
