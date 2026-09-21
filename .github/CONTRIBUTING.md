# Contributing

Thanks for taking the time. Issues and pull requests are welcome.

## Development

```bash
go test ./...                              # unit tests
go -C tools/oggen test ./...               # oggen tests (separate module)
BASE_FQDN=example.com PORT=8080 PUBLIC_DIR=example/public go run ./cmd/karako-cdn
./test.sh 8080                             # integration suite
lefthook install                           # git hooks
```

Everything in the repository is written in English, commit messages included. The server module
keeps zero external dependencies: anything needing a library belongs in
a separate module under `tools/`.

## Pull requests

- One coherent change per pull request, with tests.
- Commit subjects follow
  [Conventional Commits](https://www.conventionalcommits.org)
  (`feat:`, `fix:`, `docs:`, `ci:`, ...). They drive the version and the
  changelog, so a `feat:` or a `fix:` is what triggers a release.
- `gofmt`, `go vet`, `actionlint` and the tests must pass. The git hooks
  run them for you.
