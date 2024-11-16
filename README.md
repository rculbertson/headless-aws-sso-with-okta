# headless-aws-sso-with-okta

Run `aws sso login` without having to open a browser, or click any buttons. 

[![Go Reference](https://pkg.go.dev/badge/github.com/rculbertson/headless-aws-sso-with-okta.svg)](https://pkg.go.dev/github.com/rculbertson/headless-aws-sso-with-okta) [![Go Report Card](https://goreportcard.com/badge/github.com/rculbertson/headless-aws-sso-with-okta)](https://goreportcard.com/report/github.com/rculbertson/headless-aws-sso-with-okta)

## Install

To download the latest release, run:

```sh
 curl --silent --location https://github.com/rculbertson/headless-aws-sso-with-okta/releases/latest/download/headless-aws-sso-with-okta_0.1.0_$(uname -s)_$(uname -m).tar.gz | tar xz -C /tmp/
 sudo mv /tmp/headless-aws-sso-with-okta /usr/local/bin
```

Alternatively:

```sh
go install github.com/rculbertson/headless-aws-sso-with-okta@latest
```

**Windows**: Download latest Windows binary from the [Releases Page](https://github.com/rculbertson/headless-aws-sso-with-okta/releases) and unzip to location in PATH

## Usage:

### Authenticate with Okta Verify and push notification

```bash
aws sso login --no-browser | ./headless-aws-sso-with-okta --okta-auth push-notification
```

### Authenticate with Okta Verify and FastPass

```bash
aws sso login --no-browser | ./headless-aws-sso-with-okta --okta-auth fastpass
```

### Authenticate with email and push notification

Use this method on linux, which doesn't support Okta Verify. Must use push-notification.

```bash
aws sso login --no-browser | ./headless-aws-sso-with-okta --email <EMAIL> --okta-auth push-notification
```

**Example:**

![headless-aws-sso-with-okta demo](./docs/demo.gif)

### License:

Apache-2.0
