# Auth Reference

Use this reference when the task is about authentication, profile selection, or verifying which chill.institute environment the CLI is targeting.

## Rules

- Parse only `stdout`. Browser hints and notices may appear on `stderr`.
- Check the active host first with `chilly settings get api-base-url --output json`.
- Use `--profile <name>` or `--config <path>` when you need isolated state.
- Use `whoami` after auth changes when you need positive confirmation that the token works.
- Use `doctor --output json` when auth and local config may both be part of the problem.

## Canonical Commands

- Interactive login: `chilly auth login`
- Non-interactive existing token: `chilly auth login --token <token>`
- Localhost callback flow: `chilly auth login --local-browser`
- Localhost callback flow without auto-open: `chilly auth login --local-browser --no-browser`
- Preview token login from stdin JSON: `printf '{"token":"token-from-setup","skip_verify":true}' | chilly auth login --json @- --dry-run --output json`
- Logout: `chilly auth logout --output json`
- Preview logout: `chilly auth logout --dry-run --output json`
- Verify current auth: `chilly whoami --output json`
- Show auth and host diagnostics together: `chilly doctor --fields auth.status,config.api_base_url,config.profile --output json`

## Host And Profile Checks

- Current host only: `chilly settings get api-base-url --fields value --output json`
- Current config path: `chilly settings path --fields path --output json`
- Current profile and host: `chilly settings show --fields profile,api_base_url --output json`
- Isolated profile path: `chilly settings path --profile dev --output json`

## Browser Token Flow

`chilly auth login` now defaults to the hosted web token flow: it prints [https://chill.institute/auth/cli-token](https://chill.institute/auth/cli-token), tells the user to copy the setup token, then waits for that token to be pasted back into the terminal.

If the browser is on another machine, open the same page in a signed-in browser and copy the token into `chilly auth login --token <token>`.
