# npm Trusted Publishing

`chilly` publishes npm packages with npm Trusted Publishing from GitHub Actions. Do not add an `NPM_TOKEN` secret for the normal release path.

The root package is `@chill-institute/cli`; it installs the `chilly` binary through platform-specific optional dependencies:

- `@chill-institute/cli-darwin-arm64`
- `@chill-institute/cli-darwin-x64`
- `@chill-institute/cli-linux-arm64`
- `@chill-institute/cli-linux-x64`
- `@chill-institute/cli-win32-arm64`
- `@chill-institute/cli-win32-x64`

## GitHub Environment

The `release` GitHub Environment is used as the release trust boundary. It should allow:

- the default branch `main`
- protected release tags matching `v*`

Release jobs set `deployment: false` so npm and tap publishing can read Environment secrets and variables without creating GitHub Deployment records. Do not attach custom deployment protection rules to this Environment unless `deployment: false` is removed.

Keep `TAP_GITHUB_TOKEN` scoped to the Homebrew tap and stored on the `release` Environment. npm publishing uses OIDC instead of a static registry token.

## npm Publisher Setup

Login once with an npm owner or admin account, then register every package for both publishing workflows:

```bash
npx -y npm@^11.10.0 login

for package in \
  @chill-institute/cli \
  @chill-institute/cli-darwin-arm64 \
  @chill-institute/cli-darwin-x64 \
  @chill-institute/cli-linux-arm64 \
  @chill-institute/cli-linux-x64 \
  @chill-institute/cli-win32-arm64 \
  @chill-institute/cli-win32-x64
do
  npx -y npm@^11.10.0 trust github "$package" \
    --repo chill-institute/chill-cli \
    --file main.yml \
    --env release \
    --yes

  npx -y npm@^11.10.0 trust github "$package" \
    --repo chill-institute/chill-cli \
    --file release.yml \
    --env release \
    --yes
done
```

The `main.yml` publisher covers the normal semantic-release path. The `release.yml` publisher covers the protected tag fallback.

## Release Checks

Before npm publish, the workflow:

1. builds binaries with GoReleaser
2. generates npm package directories from GoReleaser artifacts
3. runs `npm pack --dry-run` for every generated package
4. publishes platform packages first, then `@chill-institute/cli`

Do not set `registry-url` in `actions/setup-node` for this path. Trusted Publishing authenticates with the job OIDC identity and `npm publish --provenance`.
