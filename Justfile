default:
    @just --list

# Full quality gate: format check + vet + lint + test + build.
# Single source of truth for the gates — CI and pre-commit call this same recipe.
ci: fmt-check vet lint test build

# Format all Go code in place (gofmt).
fmt:
    gofmt -w .

# Verify formatting; fail listing any file that is not gofmt-clean.
fmt-check:
    #!/usr/bin/env bash
    set -euo pipefail
    unformatted="$(gofmt -l .)"
    if [ -n "$unformatted" ]; then
        echo "not gofmt-clean:"; echo "$unformatted"; exit 1
    fi

# Vet: report suspicious constructs (go's built-in analyzer).
vet:
    go vet ./...

# Static analysis (staticcheck). No warnings tolerated.
lint:
    staticcheck ./...

# Run the test suite. Export DSA_SAVE=/path/to/<id>_Slot<N>.db to also exercise
# a real save (decrypt/edit/re-encrypt round-trips).
test:
    go test ./...

# Build every package for the host platform.
build:
    go build ./...

# Cross-compile the Windows editor: a single static .exe, no CGO.
build-windows:
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/dsa-save-editor.exe ./cmd/dsa-save-editor

# Run the editor against a save file (opens the browser UI).
run save:
    go run ./cmd/dsa-save-editor {{ save }}

# Regenerate CHANGELOG.md from the Conventional Commits (git-cliff). Review the diff.
changelog:
    git-cliff --output CHANGELOG.md

# Update Go dependencies and tidy the module graph.
update:
    go get -u ./... && go mod tidy
