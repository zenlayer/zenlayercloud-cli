# Reading `zeno --help` and parameter types

`zeno <service> <api> --help` is the source of truth for every command. Read it before invoking any API you have not just run. This reference explains how to parse the output and how to pass each parameter type.

## Help page anatomy

```
NAME
DESCRIPTION        ← what the API does; often links to prerequisite APIs you should call first
SYNOPSIS           ← flag skeleton; flags in [brackets] are optional, flags without are required
OPTIONS            ← per-flag: name, type, allowed values, defaults
OUTPUT             ← response shape (useful for building JMESPath -q queries)
EXAMPLES           ← real invocations; copy the shape, not the IDs
GLOBAL OPTIONS     ← flags shared by every command: --cli-dry-run, -o, -q, --profile, --debug...
```

Rules of thumb:

- If SYNOPSIS shows `--foo <value>` (no brackets), `--foo` is required.
- If OPTIONS describes a flag as `(list)` or `(structure)`, it is not a plain string — see below.
- If the DESCRIPTION says "call `DescribeX` first to obtain Y", do exactly that — do not guess Y.

## Parameter types

### `(string)` and `(integer)` and `(boolean)`

Plain scalars — pass as-is.

```bash
--instance-name web-01
--instance-count 2
--enable-agent true
```

Boolean flags use `true` / `false` literals, not flag presence.

### `(list)` — scalar list

Repeat the flag once per value. Example: `describe-zones --zone-ids`:

```bash
zeno zec describe-zones \
  --zone-ids asia-east-1a \
  --zone-ids asia-south-1a \
  -o json
```

Not `--zone-ids asia-east-1a,asia-south-1a` — that's a single value `"asia-east-1a,asia-south-1a"`.

### `(structure)` — single object

Two syntaxes. Both appear in the help page under **Shorthand Syntax** and **JSON Syntax**.

Shorthand (compact, no spaces in values):

```bash
--system-disk diskSize=40,diskCategory="Standard NVMe SSD"
```

JSON (required when values contain commas, equals signs, quotes, or nested objects):

```bash
--system-disk '{"diskSize":40,"diskCategory":"Standard NVMe SSD","burstingEnabled":true}'
```

Always single-quote the JSON so your shell does not eat the double quotes. Confirm in `--help` which field names are input-only vs response-only — e.g. `diskId` on `create` is response-only and must not be sent.

### `(list)` of `(structure)` — array of objects

Shorthand: repeat the flag per element, each element using `key=val,key=val`:

```bash
--tags key=env,value=prod --tags key=team,value=infra
```

JSON: pass a JSON array:

```bash
--tags '[{"key":"env","value":"prod"},{"key":"team","value":"infra"}]'
```

When in doubt, use JSON — it handles nested structures and special characters cleanly.

## Global flags you will use constantly

| Flag               | Purpose                                                                  |
|--------------------|--------------------------------------------------------------------------|
| `-o json`          | Machine-readable output. Always use for scripting.                       |
| `-q '<JMESPath>'`  | Filter/project the response server-side view; reduces noise.             |
| `--cli-dry-run`    | Print the signed request without sending it. **Required before writes.** |
| `--profile <name>` | Use a named profile instead of the default.                              |
| `--debug`          | Log the full HTTP request/response. Use once on error, then remove.      |
| `--page-all`       | Transparently fetch all pages of a paginated list.                       |
| `--endpoint <url>` | Override API domain. Only for internal / non-prod endpoints.             |

## JMESPath queries (`-q`)

JMESPath shapes JSON output into just what you need — great for piping into shell variables.

```bash
# All zone IDs as a JSON array
zeno zec describe-zones -o json -q 'zoneSet[*].zoneId'

# The first Linux image in a zone
zeno zec describe-images --zone-id asia-east-1a -o json \
  -q 'imageSet[?osType==`Linux`] | [0].imageId'

# Running instances only, just id + name
zeno zec describe-instances -o json \
  -q 'dataSet[?status==`Running`].{id:instanceId,name:instanceName}'
```

Note JMESPath uses backticks for literals (`` `Linux` ``) and single-quote the whole expression so the shell passes it through untouched.

## Extracting IDs into shell variables

```bash
# strip the outer quotes from a single string result
ZONE_ID=$(zeno zec describe-zones -o json -q 'zoneSet[0].zoneId' | tr -d '"')

# or use jq for anything nested
ZONE_ID=$(zeno zec describe-zones -o json | jq -r '.zoneSet[0].zoneId')
```

Never hand-type an ID you saw in a previous run — pass it through a variable so there's one source of truth.

## Dry-run discipline

`--cli-dry-run` is free and catches:

- missing required flags
- wrong flag names (typos)
- malformed `(structure)` / `(list)` values
- signing / credential problems (you'll see `AUTHFAILURE_SIGNATURE_FAILURE` here before any change)

Workflow:

```bash
# 1. Build the command with --cli-dry-run
zeno zec create-zec-instances <flags...> --cli-dry-run -o json

# 2. Show the user the exact command you're about to run and the concrete
#    resources it will affect. Wait for an explicit "yes".

# 3. Re-run the same line without --cli-dry-run.
```

Never copy a dry-run success and assume the real call will work — it re-signs and re-sends, which can still fail (quota, concurrent state change). Check the real response.

## Pagination

List APIs accept `--page-num` (1-based) and `--page-size` (commonly 20–100). For anything that might exceed one page, use `--page-all` to iterate for you:

```bash
zeno zec describe-instances -o json --page-all -q 'dataSet[*].instanceId'
```

Do not client-side filter huge result sets when the API supports server-side filters (e.g. `--zone-id`, `--status`, `--resource-group-id`) — it's slower and costlier.

## Debugging a failing command

1. Re-read the `--help`. 90% of failures are wrong flag name, wrong enum value, or a required flag omitted.
2. Add `--cli-dry-run` — does signing succeed?
3. Add `--debug` — read the actual request body and error message. Remove once fixed.
4. Check the error code against `docs.console.zenlayer.com/zenlayer-cli` and the service-specific API reference.

Do not retry the same command with a blind tweak. Each retry on a write path is a potential side effect.
