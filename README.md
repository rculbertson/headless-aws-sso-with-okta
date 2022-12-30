## headless-aws-sso-with-okta
Runs [aws sso login]() headlessly when using the `--no-browser` option.

[![Go Reference](https://pkg.go.dev/badge/github.com/ekesken/headless-aws-sso-with-okta.svg)](https://pkg.go.dev/github.com/ekesken/headless-aws-sso-with-okta) [![Go Report Card](https://goreportcard.com/badge/github.com/ekesken/headless-aws-sso-with-okta)](https://goreportcard.com/report/github.com/ekesken/headless-aws-sso-with-okta)

### Background

We want to avoid leaving the terminal and opening yet another tab and having to click Next next next...

### Install

To download the latest release, run:
> For ARM systems, please change ARCH to `arm64`

``` sh
 curl --silent --location https://github.com/ekesken/headless-aws-sso-with-okta/releases/latest/download/headless-aws-sso-with-okta_0.1.0_$(uname -s)_x86_64.tar.gz | tar xz -C /tmp/
 sudo mv /tmp/headless-aws-sso-with-okta /usr/local/bin
```

Alternatively:

``` sh
go install github.com/ekesken/headless-aws-sso-with-okta@latest
```

**Windows**: Download latest Windows binary from the [Releases Page](https://github.com/ekesken/headless-aws-sso-with-okta/releases) and unzip to location in PATH

#### Dependancies:

This tool requires you to have installed and configured Dashlane's official CLI tool. You can find installation instructions here: [Dashlane CLI](https://github.com/Dashlane/dashlane-cli/tree/master/src)
### Usage:

``` bash
aws sso login --profile login --no-browser | ./headless-aws-sso-with-okta "${USERNAME}" '${PASSWORD}' "$(oathtool -b --totp ${MFA_SECRET})"
```

**Example:**

![headless-aws-sso-with-okta demo](./docs/demo.gif)

### Release Notes:
Working but still WiP, contributions welcome.

### License:
Apache-2.0
