# CLAUDE.md — Harbor CLI architecture & contributor guide

This document orients a developer (human or AI) working on the Harbor CLI
codebase: how it is laid out, the patterns to follow, how to build and test, and
a step-by-step recipe for adding a new command. It is intentionally
self-contained — you do not need any private/internal knowledge to contribute.

## What this is

`harbor` is a Go + [cobra](https://github.com/spf13/cobra) command-line client
for the public Harbor notes API. It speaks the API's JSON wire contract and
renders results as styled tables (via [go-pretty](https://github.com/jedib0t/go-pretty))
or raw JSON (`--json`). It is a thin, well-tested client: business logic lives
in the API; the CLI's job is ergonomics, output, and auth.

## Layout

```
main.go                 → cmd.Execute()
cmd/                    cobra command tree — ONE FILE PER DOMAIN
  root.go               rootCmd, global flags, loadClientFromConfig, printResult
  display.go            color, JSON-navigation + formatting helpers, tables, error rendering
  flags.go              shared flag helpers (paging, partial-update body building)
  input.go              body input (--content/--file/--stdin) + flexible time parsing
  auth.go               login, logout, whoami, auth refresh, public recovery flows
  notebooks.go          notes.go  tags.go  note_tags.go  sync.go  files.go  search.go
  history.go  trash.go  templates.go  shortcuts.go  reminders.go  insights.go
  share.go  settings.go  profile.go  sessions.go  account.go  importexport.go  operational.go
client/                 HTTP client + API methods — ONE FILE PER DOMAIN
  client.go             Client, HTTP verbs, transparent-refresh, request plumbing
  errors.go             APIError (typed error envelope)
  envelope.go           response-envelope decoders (collection/paging/data/token)
  auth.go  notebooks.go  notes.go  …                (mirror cmd/ domains)
config/
  config.go             credentials.json load/save (0600), expiry helpers
Formula/harbor.rb       Homebrew formula (version auto-bumped by release CI)
.github/workflows/      test.yml (CI) + release.yml (CD)
```

Every source file `X.go` has a sibling `X_test.go` (one test file per file).

## The wire contract (what the client speaks)

- **Base URL** includes the `/api/v1` prefix; client method paths are relative
  (`/notes`, `/notebooks/:id`). The few **operational** probes (`/health`,
  `/ready`, `/version`) live at the **root** — derive it with `c.Origin()`.
- **JSON** request/response bodies; field names are `snake_case`.
- **Timestamps** are UTC epoch-milliseconds (integers). The OAuth `expires_in`
  is the lone exception (seconds).
- **Response envelopes** (see `client/envelope.go`):
  - collection: `{ "data": [...], "paging": { limit, offset, total, has_more } }`
  - wrapped single: `{ "data": {...} }`
  - bare single: many create/get/update endpoints return the resource directly
  - note mutations: `{ "note": {...}, "usn": N }`
  - OAuth token: a bare `{ access_token, refresh_token, … }`
- **Errors**: `{ "error": { code, message, details, request_id } }`, decoded into
  `*client.APIError`. Commands branch on `Code`.

## Core patterns (copy these)

**A command's `RunE`** loads a client, builds a request from flags, calls a
client method, and prints the result:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    c, _, err := loadClientFromConfig()
    if err != nil {
        return err
    }
    data, err := c.ListNotebooks(pagingParams(cmd))
    if err != nil {
        return mapNotebookError(err) // friendly per-domain messages
    }
    printResult(data, displayNotebooks) // JSON in --json mode, else the table
    return nil
}
```

- **Never `os.Exit` inside `RunE`** — return the error; `Execute` renders it
  (rich treatment for `*client.APIError`) and sets the exit code.
- **`printResult(data, displayFn)`** handles `--json` automatically; your
  `displayFn(data []byte)` only renders the table/detail view.
- **Partial updates**: only send fields the user set, via
  `addStringIfChanged` / `addBoolIfChanged` / `addIntIfChanged`.
- **Pagination**: `addPagingFlags(cmd)` + `pagingParams(cmd)` for consistent
  `--limit/--offset/--order` everywhere; call `printPagingFooter(data)` after a
  list table.
- **Public endpoints** (login, password reset, public share, ops) use
  `newAnonymousClient()` (no bearer); everything else uses
  `loadClientFromConfig()`.
- **Friendly errors**: add a `map<Domain>Error(err) error` that switches on the
  domain's `APIError.Code` values and returns `errors.New(...)`; let everything
  else fall through to the default renderer (which prints `details`).
- **Self-registration**: each domain file's `init()` calls
  `rootCmd.AddCommand(...)` (or attaches to an existing parent like `notesCmd`).
  Set a `GroupID` (`groupAuth`/`groupContent`/`groupOrg`/`groupSync`/
  `groupAccount`/`groupSystem`) so `--help` groups it.

## Add a new endpoint (recipe)

Say the API gains `GET /api/v1/widgets` and `POST /api/v1/widgets`.

1. **Client methods** — `client/widgets.go`:
   ```go
   func (c *Client) ListWidgets(params map[string]string) ([]byte, error) {
       return c.doGet("/widgets", params)
   }
   func (c *Client) CreateWidget(body map[string]any) ([]byte, error) {
       return c.doPost("/widgets", body)
   }
   ```
   Use `doGet/doPost/doPatch/doPut/doDelete`, `doMultipart` (uploads), or
   `doGetRaw` (streamed downloads). Need the HTTP status (e.g. 201-vs-200)?
   mirror `AttachTag` in `client/note_tags.go` (`requestWithStatus`).

2. **Commands** — `cmd/widgets.go`: a parent `widgetsCmd` + subcommands, a
   `displayWidgets`/`displayWidget` renderer using `printTable`/`printKV` and the
   shared helpers in `display.go`, a `mapWidgetError`, and an `init()` that wires
   flags and `rootCmd.AddCommand(widgetsCmd)`.

3. **Tests** — `client/widgets_test.go` asserts method/path/query/body against a
   mock server (`newTestServer` + `testClient`); `cmd/widgets_test.go` feeds
   fixture JSON to the display funcs via `captureStdout` and checks the error
   mapping with `apiErr`.

4. Put the Cloudmanic copyright header (current year + date) atop each new file,
   and a doc comment above every function. Run `make lint && make test`.

## Build & test

```sh
make build          # ./build/harbor, version injected via ldflags
make test           # go test ./... (no network, no config required)
make lint           # gofmt + go vet
make cross-build    # all release platforms into dist/
make run ARGS="notes list --json"
```

`go test ./...` MUST pass with **no network and no config**. Tests mock the API
with `httptest.NewServer` and use a temp `HOME` for config tests, so they never
touch a real server or your real `~/.config/harbor`.

### Testing recipe

- **Client tests** use the shared harness in `client/client_test.go`:
  `newTestServer(t, &rec, status, body)` records the request into a
  `recordedRequest` (method/path/query/body/headers); `testClient(url)` points a
  `Client` at it. Assert the verb, path, query, and decoded JSON body.
- **Command tests** use `captureStdout(t, fn)` (in `cmd/display_test.go`) to
  capture a display function's output, and `apiErr(code)` to build an
  `*client.APIError` for `map<Domain>Error` tests.
- **Config tests** call `t.Setenv("HOME", t.TempDir())` so the credentials file
  is isolated.
- Use obviously-fake fixture data (`you@example.com`, ids like `n1`).

## Credentials file

`~/.config/harbor/credentials.json` (mode `0600`), written atomically:

```json
{
  "api_url": "https://app.harbor.my/api/v1",
  "client_id": "harbor-app",
  "email": "you@example.com",
  "user_id": "…",
  "access_token": "at_…",
  "refresh_token": "rt_…",
  "token_type": "Bearer",
  "scope": "notes notebooks tags sync files search profile",
  "expires_at": 1750000000000,
  "device_id": "cli-…",
  "device_name": "harbor-cli on <host>"
}
```

`api_url` is omitted (defaults to production) for normal logins; it is set only
when a maintainer targets a non-default environment. Tokens are never logged or
printed (use `--show-token` to opt in explicitly).

## Releases

Pushing to `main` triggers `release.yml`: it tests, auto-increments the patch
version off the latest `vX.Y.Z` tag (first release `v0.1.0`), cross-compiles the
five platform binaries with the version baked in via ldflags, publishes a GitHub
release, and bumps `Formula/harbor.rb` so `brew upgrade harbor` works.
