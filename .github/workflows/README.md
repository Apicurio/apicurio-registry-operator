# Release Process

The Apicurio Registry Operator release process consists of 5 phases, most of which are automated, but require some manual preparation and possibly intervention.

## Phase 0: Preparation

Each release involves two versions:
 - Version of the operator, e.g. `1.1.3`,
 - Version of the operand, e.g. `2.6.4.Final`.

The CSV version is a combination of these, e.g. `1.1.3-v2.6.4.final`.

*NOTE: At the moment, the workflows assume that you are releasing new version of both components. Until a lightweight workflow that only updates the operand version is added, you should simply increment the operator version even if there are no changes.*

Add and commit release notes under the new CSV version. See e.g. `docs/resources/release-notes/1.1.3-v2.6.4.final.md` for reference.

## Phase 1: Build

Run workflow named **Release Phase 1 - Prepare release branch, build, test, and push images**.

Example arguments:
 - Use workflow from: `Branch: main`
 - Operator version being released: `1.1.3`
 - Operand version to use: `2.6.4.Final`
 - Branch to release from: `main`
 - Branch used during the release: `release`
 - Debug with tmate on failure: `true`

This workflow is responsible for:
 - updating operator and operand versions
 - building and testing the executable
 - building operator, bundle, and catalog images
 - pushing the images
 - building dist archive (used for GH release)
 - running e2e tests
 - pushing the `release` branch that is used for the next phase

What can go wrong:
 - If the workflow fails, you can connect to the runner using ssh and try to debug
 - If the envtests fail, you can run the tests locally with `make build`.
 - If the e2e tests fail, you can edit the e2e tests repository https://github.com/Apicurio/apicurio-registry-k8s-tests-e2e.git and try to debug.
 - Make sure the credentials are valid in case push to quay fails.

After the workflow runs successfully, there should be a `release` branch available in the operator repository. This is done in case phase 2 requires human intervention, so user can edit e.g. bundle files. However, it may be best to
fix the issue and run phase 1 again, since we also build bundle images (although not used in Operator Hub and Openshift Marketplace).

## Phase 2: Create Operator Hub PRs

Run workflow named **Release Phase 2 - Create Operator Hub PRs**.

Example arguments:
- Use workflow from: `Branch: main`
- Branch used during the release: `release`
- Debug with tmate on failure: `true`

This workflow is responsible for opening two PRs in the Operator Hub and Openshift Marketplace repos, e.g.:
- https://github.com/k8s-operatorhub/community-operators/pull/5038
- https://github.com/redhat-openshift-ecosystem/community-operators-prod/pull/5191

The links to the PRs are listed in the workflow logs.

Check the PRs, make sure the commits look good, and wait for the automated checks and tests to pass. Go through the checkboxes and check them as appropriate.

What can go wrong:
 - The workflow for OpenShift community operators will likely fail with `check_operator_name_unique` error. For now, wait or contact one of the reviewers and ask for `tests/skip/check_operator_name_unique` label to be added (See the example).
 - The workflow is using credentials of `jsenko` to open the PRs. If you need to do any changes, it's possible that you need to run the workflow under your name. Edit the workflow and add your GH token with `repo` and `workflow` permissions. The list of allowed users is here https://github.com/redhat-openshift-ecosystem/community-operators-prod/blob/main/operators/apicurio-registry/ci.yaml .

## Phase 3: Build and publish the Java API Model

Run workflow named **Release Phase 3 - Build and publish the Java API Model**.

Example arguments:
- Use workflow from: `Branch: main`
- Branch used during the release: `release`
- Debug with tmate on failure: `true`

This phase is responsible for building and releasing the Java operator model.

What can go wrong:
- Sometimes the Sonatype staging repository is flaky and the deployment might fail. Retrying some time later usually works. 

## Phase 4: Create GH release

Run workflow named **Release Phase 4 - Tag, create a release in GitHub, cleanup and set up next dev version**.

Example arguments:
- Use workflow from: `Branch: main`
- Next Operator version (-dev): `1.2.0-dev`
- Next Operand version: `2.6.x`
- Next Operand image version: `2.6.x-snapshot`
- Branch used during the release: `release`
- Debug with tmate on failure: `true`

This phase is responsible for finalizing the release, creating tags and GH release.

What can go wrong:
- This last phase is usually without issues, and anything should be fixable manually with `git` and in GitHub.
