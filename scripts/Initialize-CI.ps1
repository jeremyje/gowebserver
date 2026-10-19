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

# Per-project CI preparation hook for Windows jobs.
#
# The CI workflow runs this immediately before "make system-info" in the
# build-windows job (see .github/workflows/deploy.yaml), so it is the place
# to install or configure whatever a specific project needs on its Windows
# runners: Windows features, extra toolchain, runner-local fixes.
#
# It deliberately does nothing here. scripts/template-go/sync-template.sh
# copies this file into a project only when the project does not already have
# one and never overwrites an existing copy, so anything added below survives
# a template re-sync. A non-zero exit status fails the job, which is the intended way to
# make a prerequisite failure visible.

$ErrorActionPreference = 'Stop'
