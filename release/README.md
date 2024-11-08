# Release Pipelines

This directory contains the Tekton Pipelines used to release EC from the main branch. These
Pipelines execute in [Konflux](https://konflux-ci.dev/).

The Pipelines are generated via [kustomize](https://kustomize.io/) from the `src` directory. To
make changes to the Pipelines, update the corresponding files in that directory and run the
`make generate-pipelines` command (requires `kustomize`).

## Setup

The [setup.yaml](setup.yaml) file should be applied to the namespace where the release Pipeliens
will run. This creates a ServiceAccount with access to perform the release.

## Why are there two Pipelines?

Currently, it is not possible to specify the EC policy in the ReleasePlan, nor any general Pipeline
parameter. Because the CLI and the Tekton Task require different EC policies, the only way to
achieve this is by using different Pipelines with different default values for the EC policy.

## Hacking

Build Pipeline

- Build CLI image as usual
- Build bundle image
- TODO: Use referrer's API to connect them. Maybe...

Release Pipeline

- Fetch Snaphost. It should have a single component, the ec-cli.
  Emit this as a result.
- From the ec-cli image, download attestation, and get the tekton bundle image reference.
  A bit of hard-coding here. That's ok.
  For now, just use cosign download att. Image will be validated by EC anyways.
- Generate a new snapshot that contains the bundle image.
  Use the same source info for the tekton bundle.
  Emit this as a result. (Size should be ok since it's a single component snapshot.)
- Run 2 EC tasks, one for each snapshot.
- Next just push each image to the expected location.
  CLI first, then bundle. One TaskRun.

- TODO: Add explicit policies to verify the bundle image? `images:` should be from a trusted registry
  or come from the same snapshot.
