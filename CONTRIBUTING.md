# Contributing

Thanks for your interest in contributing to stripe-payout-reconciler!

## Prerequisites

- Go 1.23 or later
- A Stripe test-mode API key (for integration testing)
- [golangci-lint](https://golangci-lint.run/) (for linting)

## Development Setup

```sh
git clone https://github.com/paco/stripe-payout-reconciler.git
cd stripe-payout-reconciler
make build
make test
```

## Running Tests

```sh
make test      # runs go test -race ./...
make vet       # runs go vet ./...
make lint      # runs golangci-lint
```

## Code Style

- Run `make fmt` before committing
- Follow existing patterns — look at neighboring code for conventions
- golangci-lint enforces the project's style rules (see `.golangci.yml`)

## Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scope): add new feature
fix(scope): fix a bug
docs: update documentation
test: add or update tests
chore: maintenance tasks
```

## Pull Request Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Make your changes with tests
4. Ensure CI passes: `make test && make vet && make lint`
5. Open a pull request against `main`

## Security

If you find a security vulnerability, **do not open a public issue**. See [SECURITY.md](SECURITY.md) for responsible disclosure instructions.
