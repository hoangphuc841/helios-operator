# Git Hooks for Helios Operator

This directory contains Git hooks to ensure code quality and consistency.

## Available Hooks

### pre-commit

Runs automatically before each commit to ensure code quality:

1. **Go Formatting**: Checks and auto-formats Go code with `gofmt`
2. **Go Vet**: Runs `go vet` to find suspicious constructs
3. **Linting**: Runs `golangci-lint` if available
4. **Unit Tests**: Runs unit tests with race detection
5. **Manifest Generation**: Ensures CRD manifests are up-to-date
6. **Common Issues**: Checks for TODO/FIXME comments and fmt.Println usage

## Setup

### Install Hooks

Run the setup script from the project root:

```bash
./.githooks/setup.sh
```

Or manually configure:

```bash
git config core.hooksPath .githooks
```

### Verify Setup

```bash
git config core.hooksPath
# Should output: .githooks
```

## Usage

### Normal Commits

Just commit as usual. The hooks will run automatically:

```bash
git add .
git commit -m "feat: add new feature"
```

If any check fails, the commit will be blocked and you'll see which checks failed.

### Skip Hooks (Not Recommended)

In rare cases where you need to bypass hooks:

```bash
git commit --no-verify -m "wip: work in progress"
```

⚠️ **Warning**: Only use `--no-verify` when absolutely necessary. All code should pass checks before being committed.

## Disable Hooks

To completely disable the hooks:

```bash
git config --unset core.hooksPath
```

## Troubleshooting

### Hook Not Running

1. Verify hooks path is set:

   ```bash
   git config core.hooksPath
   ```

2. Ensure script is executable:

   ```bash
   chmod +x .githooks/pre-commit
   ```

3. Re-run setup:
   ```bash
   ./.githooks/setup.sh
   ```

### Hook Fails on golangci-lint

If you don't have golangci-lint installed:

```bash
# Install globally
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Or use project's Makefile
make golangci-lint
```

### Hook Fails on controller-gen

Install required tools:

```bash
make tools
```

### Hook Takes Too Long

For quick commits during development, you can temporarily disable specific checks by commenting them out in `.githooks/pre-commit`, or use `--no-verify` sparingly.

## Customization

You can modify `.githooks/pre-commit` to:

- Add/remove checks
- Adjust timeout values
- Change warning vs. error behavior
- Add project-specific validations

## Best Practices

1. **Run checks locally before committing**: `make verify`
2. **Keep commits atomic**: Smaller commits are easier to verify
3. **Fix issues promptly**: Don't let TODOs accumulate
4. **Use structured logging**: Avoid fmt.Println in production code
5. **Keep manifests updated**: Run `make manifests generate` after API changes

## Integration with CI/CD

These same checks should run in CI/CD pipelines. The pre-commit hook provides fast feedback locally, while CI ensures nothing slips through.

See `.github/workflows/ci.yaml` for the CI configuration.
