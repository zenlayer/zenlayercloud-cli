---
name: zeno-guidance
description: Use when managing Zenlayer Cloud resources (BMC bare metal, ZEC elastic compute, ZLB load balancers, ZOS ororchestration system, ZRM resource management, CCS(key pair), IPT (IP transit), traffic, USER (resource group & permission management)) or writing shell/Python scripts that call Zenlayer. Also use whenever the user mentions `zeno`, `zenlayer-cli`, "Zenlayer Cloud", Zenlayer API error codes (e.g. AUTHFAILURE_SIGNATURE_FAILURE), asks to list/create/modify/delete Zenlayer resources, asks how to authenticate to Zenlayer, or pastes a command starting with `zeno`. Covers install/credential/profile setup, discovering commands via `--help`, parameter typing (list/structure shorthand vs JSON), dry-run validation, confirmation gates for destructive ops, and JSON-only output for scripting.
---

# Zenlayer Cloud CLI (zeno)

`zeno` is the official CLI for Zenlayer Cloud. It is the **only** sanctioned interface for this skill to touch Zenlayer — it signs requests, manages profiles, and enforces safety flags that raw HTTP clients do not.

Official docs: https://docs.console.zenlayer.com/zenlayer-cli

## MUST NOT

1. **Do not use any channel other than `zeno` to operate Zenlayer resources.** No `curl`, no direct HTTP calls to `*.zenlayer.com` APIs, no browser automation against the console, no third-party SDKs. Zenlayer's API requires HMAC request signing; ad-hoc HTTP calls will either fail with `AUTHFAILURE_SIGNATURE_FAILURE` or — worse — leak credentials through argv/history. If `zeno` lacks a capability, say so and stop.
2. **Do not fabricate resource identifiers.** Instance IDs, zone IDs, image IDs, subnet IDs, EIP IDs, VPC IDs, disk IDs, snapshot IDs, security group IDs, key pair IDs, etc. must come from the JSON response of a prior `zeno describe-*` / `zeno create-*` call in this session. If you do not have it, run the relevant describe command first and extract from the result (see "Chaining commands" below). Never copy an ID from documentation examples and treat it as real.
3. **Do not guess flag names, subcommand names, or enum values.** Zenlayer commands have non-obvious naming (`create-zec-instances`, not `create-instance`; `instance-type`, not `instance-size`; zone IDs like `asia-east-1a`). Before constructing any command you have not just run, call `zeno <service> <api> --help` and read the actual flag names and the accepted values. See `references/command-types.md` for how to parse the help output.
4. **Do not print credentials in plaintext.** Never `cat ~/.zenlayer/credentials.json`, never echo `$ZENLAYER_ACCESS_KEY_SECRET`, never include a real access key/secret in code, tests, commit messages, or chat output. If you must show a value, redact to the last 4 characters (e.g. `****ab12`). Do not commit `~/.zenlayer/credentials.json` or any file containing real keys to git.

## MUST

1. **Confirm intent before any write or delete operation.** Any command whose verb is `create-*`, `modify-*`, `update-*`, `delete-*`, `terminate-*`, `release-*`, `apply-*`, `assign-*`, `associate-*`, `attach-*`, `detach-*`, `reboot-*`, `start-*`, `stop-*`, `configure-*`, `cancel-*`, `change-*`, `copy-*` requires explicit user confirmation showing (a) the exact command you're about to run and (b) the concrete resources it will touch (IDs, counts, regions). Wait for a clear "yes" before executing. Read-only `describe-*` / `inquiry-*` / `available-*` calls do not need confirmation.
2. **Dry-run write operations first.** Append `--cli-dry-run` to every write/delete command as a preflight check. This previews the signed request without sending it and surfaces missing/invalid parameters. Only after the dry-run succeeds and the user confirms, run the real command (same line, minus the flag).
3. **Always emit JSON.** Use `-o json` (or the global default when already set to json) on every command. JSON is machine-parseable, stable across versions, and safe to pipe to `jq`. Table output is for humans and breaks scripts. Combine with `-q '<JMESPath>'` when you only need a slice (e.g. `-q 'dataSet[*].instanceId'`). For multi-page results use `--page-all`.

## Prerequisites

Before the first `zeno` call in a session, verify `zeno` is installed and a profile is configured:

```bash
zeno version
```

- **`zeno: command not found`** → `zeno` is not installed. Show the user `references/installation-guide.md` (install steps, credential setup, profile examples) and stop until installation is confirmed. Do not attempt to install it for them without asking.
- **Prints a version** (e.g. `1.0.2`) → proceed.

Then verify an active profile exists:

```bash
zeno configure list
```

This returns the non-secret settings (`language`, `output`, `profile`) for the active profile. It never prints the access key or secret — safe to run. If it errors with "no profile" or similar, point the user at the "Configure credentials" section of `references/installation-guide.md`.

If the user needs a different profile, use `--profile <name>` on each command or `export ZENLAYER_PROFILE=<name>` for the session. Never switch profiles silently — tell the user which profile you are using.

## Workflow

### 1. Discover the command

Zenlayer groups APIs by service. Top-level services include `bmc` (bare metal), `zec` (elastic compute), `zlb` (load balancer), `zos` (object storage), `zrm`, `ccs`, `ipt`, `traffic`.

```bash
zeno --help             # list services
zeno zec --help         # list APIs under a service (long — pipe to grep/less)
zeno zec describe-zones --help   # exact flags, types, and examples for one API
```

Prefer `zeno <service> --help | grep -i <keyword>` over guessing. The API list is flat per service, so a keyword like `subnet` or `eip` quickly narrows it.

### 2. Read the help before constructing the command

Each API's `--help` output has four sections you care about:

- **SYNOPSIS** — required vs optional flags (required have no `[...]` brackets)
- **OPTIONS** — each flag's type (`string`, `integer`, `boolean`, `list`, `structure`) and allowed values
- **EXAMPLES** — real invocations; copy the shape, not the IDs
- **GLOBAL OPTIONS** — `--cli-dry-run`, `-o`, `-q`, `--profile`, `--debug`, etc.

Pay particular attention to `(list)` and `(structure)` parameters — they accept either shorthand (`key=val,key=val`) or JSON. See `references/command-types.md` for the rules.

### 3. Chain commands: describe → extract → act

Because you must not invent IDs, typical flows look like:

```bash
# Find zone IDs
zeno zec describe-zones -o json -q 'zoneSet[*].zoneId'

# Find an image ID for that zone
zeno zec describe-images --zone-id asia-east-1a -o json \
  -q 'imageSet[?osType==`Linux`].imageId | [0]'

# Preview the create (dry-run)
zeno zec create-zec-instances \
  --zone-id asia-east-1a \
  --image-id <id-from-previous-step> \
  --instance-type <type-from-describe-zone-instance-config-infos> \
  --subnet-id <id-from-describe-subnets> \
  --instance-count 1 \
  --cli-dry-run -o json

# After user confirms, run without --cli-dry-run
```

Keep intermediate IDs in shell variables or a local scratch file — never retype from memory.

### 4. Handle errors

- `AUTHFAILURE_SIGNATURE_FAILURE` → credentials are wrong, clock is skewed, or something mutated the request after signing. Re-check the profile (`zeno configure list`), system time, and that no proxy is rewriting the request.
- `InvalidParameter` / `ResourceNotFound` → re-read the relevant `--help`; the flag name or value enum is almost always the issue. Do not retry with a guess.
- Rate-limit or transient errors → back off, don't hammer.
- For anything else, re-run with `--debug` once (it logs the request/response) and read the actual message before proposing a fix.

## Test examples

Use these as quick smoke tests after install/configure or when validating the skill is wired up. All are read-only and safe.

```bash
# 1. Installed & configured?
zeno version
zeno configure list

# 2. List zones for ZEC (elastic compute)
zeno zec describe-zones -o json

# 3. List BMC availability zones
zeno bmc describe-zones -o json

# 4. Extract just the zone IDs with JMESPath
zeno zec describe-zones -o json -q 'zoneSet[*].zoneId'
```

If `describe-zones` returns a `zoneSet`, credentials and signing are working end-to-end.

## Scripting patterns

- Set `ZENLAYER_OUTPUT=json` once at the top of a shell script instead of threading `-o json` through every call.
- Capture IDs into variables: `ZONE_ID=$(zeno zec describe-zones -o json -q 'zoneSet[0].zoneId' | tr -d '"')`.
- For idempotent lookups, prefer describe-with-filters over listing-and-filtering client-side — it's faster and paginates server-side.
- Use `--page-all` when a list may exceed a single page.
- Never interpolate user input into command flags without validation; zeno flags accept arbitrary strings and some (like `--user-data`) reach the guest VM.

## Further reading

- `references/installation-guide.md` — install, credential setup, profiles, environment variables
- `references/command-types.md` — how to read `--help`, parameter types, shorthand vs JSON, JMESPath queries
- Official reference: https://docs.console.zenlayer.com/zenlayer-cli
