# Publishing Homebrew and Scoop packages

The `Release packages` workflow builds release assets, verifies their checksums,
then commits `Formula/eitri.rb` and `bucket/eitri.json` to this repository's
`main` branch. Users install and upgrade through Homebrew or Scoop. No separate
tap/bucket repository, personal access token, or package-registry account is needed.

Homebrew builds the source package on macOS/Linux and manages Go as a build
dependency. Scoop installs the Windows x64 or ARM64 package. Users still need
the compiler/interpreter configured for their own solutions.

## First publication

1. Commit and push the release setup to `main`.
2. Enable GitHub Actions for this repository. The workflow requests
   `contents: write` for its normal `GITHUB_TOKEN`.
3. Ensure branch rules allow the release bot to update the two generated files
   on `main`. If protected branches require pull requests, use the recovery
   instructions below to submit those generated changes through your normal PR
   process instead of granting a bypass.
4. Wait for CI to pass, then create a new stable tag. `v1.0.0` already exists;
   for example, publish `v1.0.1`:

   ```sh
   git tag -a v1.0.1 -m 'Release v1.0.1'
   git push origin v1.0.1
   ```

5. Wait for **Release packages** to finish and confirm both generated package
   definitions are on `main`. The installation commands in README work after
   this first publication, not before it.

Use a new tag for each release; never move a published tag. Prerelease tags are
rejected by this stable-package workflow. Release versions are embedded in the
executable so `eitri --version` agrees with the package manager.

## What the workflow publishes

- A checksummed source archive for Homebrew, including the embedded templates.
- Windows x64 and ARM64 ZIP packages for Scoop, including LICENSE and README.
- `checksums.txt`, which is also used to generate the package definitions.
- Homebrew/Scoop definitions pointing at that exact version's release assets.

Package generation verifies the local SHA-256 of every required asset against
`checksums.txt` before writing either definition. It does not use moving `latest`
asset URLs or placeholder hashes. Package definitions change only after assets
have been published. If another commit lands on `main` during the release, the
workflow rebases its package-only commit rather than overwriting source changes.

## Local validation without publishing

Install GoReleaser v2.19.0 and run:

```sh
go test ./...
go vet ./...
go test -race ./...
python3 -m unittest discover -s scripts -p 'test_*.py'
goreleaser check
goreleaser release --snapshot --clean
```

Snapshots do not publish assets or update package definitions. To validate
package definitions against a stable tag in a clean checkout, run:

```sh
goreleaser release --skip=publish --clean
python3 scripts/package_managers.py --tag v1.0.1
```

Generated definitions appear under `dist/package-managers/`. Inspect them before
publishing. The generator deliberately rejects snapshot/prerelease versions.

## Recovering a package-definition update

If the release assets were published but the final commit failed (for example,
because `main` requires PRs), download that release's source archive, both Windows
ZIPs, and `checksums.txt` into `dist/`. Then run the generator with the release tag:

```sh
python3 scripts/package_managers.py --tag v1.0.1
```

Copy the generated `Formula/` and `bucket/` files into the repository and submit
them through the normal commit/PR process. Do not rerun the entire release to
repair a package-only failure or replace already published assets. The workflow
also accepts an existing stable tag through manual dispatch for failures that
occurred before publication.

These packages use a project-owned tap/bucket. They are not submissions to
`homebrew/core` or Scoop's `main` bucket.
