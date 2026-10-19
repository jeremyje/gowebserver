#!/usr/bin/env bash
#
# Copyright 2026 Cloudfra
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Filter a comma-separated list of git pathspecs down to the ones that match
# at least one file in the current checkout, and print the survivors in the
# same comma-separated form.
#
# peter-evans/create-pull-request stages its add-paths with a single
# `git add -- <pathspec>...`, and git rejects the whole command when any one
# pathspec matches nothing (e.g. install/terraform/*.tf in a project without
# terraform). Nothing is staged and the action fails with an empty commit.
# Running the list through this script first keeps add-paths valid in every
# project created from the template.
#
# A pathspec counts as matching when it names a tracked file (including one
# the working tree has deleted) or an untracked, non-ignored file - the same
# set `git add` would stage from it.
#
# Usage: scripts/template-go/existing-pathspecs.sh PATHSPEC[,PATHSPEC...]
#
# The result also goes to "paths" in $GITHUB_OUTPUT when that is set, for
# use as ${{ steps.<id>.outputs.paths }}. It exits non-zero when no pathspec
# matches: an empty add-paths makes create-pull-request stage everything.

set -euo pipefail

if [[ $# -ne 1 || -z "$1" ]]; then
  echo "Usage: $0 PATHSPEC[,PATHSPEC...]" >&2
  exit 1
fi

IFS=',' read -r -a candidates <<<"$1"

kept=()
for p in "${candidates[@]}"; do
  [[ -n "$p" ]] || continue
  if [[ -n "$(git ls-files --cached --others --exclude-standard -- "$p" | head -n 1)" ]]; then
    kept+=("$p")
  else
    echo "Skipping pathspec with no matching files: $p" >&2
  fi
done

if (( ${#kept[@]} == 0 )); then
  echo "None of the pathspecs match a file: $1" >&2
  exit 1
fi

result="$(IFS=','; echo "${kept[*]}")"
echo "$result"
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  echo "paths=$result" >>"$GITHUB_OUTPUT"
fi
