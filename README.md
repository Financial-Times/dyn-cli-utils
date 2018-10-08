# Dyn CLI utils

![CircleCI](https://img.shields.io/circleci/circleci/token/a3feb6b4515035e645d758ac4ddd676e1f5d2da3/project/github/Financial-Times/dyn-cli-utils/master.svg)

Utilities for interacting with the dynect API.

To install run:

```shell
go get github.com/Financial-Times/dyn-cli-utils
```

To view the available commands run:

```shell
dyn-cli-utils --help
```

or view the generated documentation at [docs](./docs/dyn-cli-utils.md).

## Configuration

Configuration is accepted as either command line flags (e.g. `--log-level`), environment variables (e.g `LOG_LEVEL), or as a config file located at either`$HOME/.dyn-cli-utils`or the given`config` path.

If a config file is provided, any of the following formats are supported: JSON, TOML, YAML, HCL, and Java properties.

## Development

Run `make help` to see available Make commands.

After making a CLI API change, please run `make docs` to regenerate static markdown documentation. These changes should be commited.

To run CLI commands, run

```shell
go run main.go [command]
```

or install the package to your `$GOBIN` directory with `go install` and then use `dyn-cli-utils` as above.
