# headless-aws-sso-with-okta

[![CI](https://github.com/rculbertson/headless-aws-sso-with-okta/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/rculbertson/headless-aws-sso-with-okta/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rculbertson/headless-aws-sso-with-okta)](https://goreportcard.com/report/github.com/rculbertson/headless-aws-sso-with-okta)

Run `aws sso login` without having to open a browser.

<img src="./docs/demo.gif" alt="Description" width="400" height="100">

## Install

### Prerequisites

On Linux, ensure nss is installed.

```bash
# Debian/Ubuntu
sudo apt-get install libnss3

# Fedora
sudo dnf install nss
```

### Linux / MacOS


```bash
curl -sL "https://github.com/rculbertson/headless-aws-sso-with-okta/releases/download/0.1.3/headless-aws-sso-with-okta_.0.1.3_$(uname -s)_$(uname -m).tar.gz" | tar xz -C /tmp/
sudo mv /tmp/headless-aws-sso-with-okta /usr/local/bin
```

### Windows

Download latest Windows binary from the [Releases Page](https://github.com/rculbertson/headless-aws-sso-with-okta/releases) and unzip to location in PATH

### From Source

```bash
go install github.com/rculbertson/headless-aws-sso-with-okta@latest
```

## Usage

The first run requires you to authenticate using either a push notification or Okta FastPass. Subsequent runs will authenticate using a saved cookie, making them much faster.

### Authenticate with email and a push notification (Linux, MacOS, Windows)

Okta Verify and FastPass do not have linux support. You must specify your email address and authenticate with a push notification.

```bash
aws sso login --no-browser | headless-aws-sso-with-okta --email <EMAIL> --okta-auth push-notification
```

### Authenticate with Okta Verify and FastPass (MacOS, Windows)

Uses FastPass's authentication method, e.g. your fingerprint.

```bash
aws sso login --no-browser | headless-aws-sso-with-okta --okta-auth fastpass
```

### Authenticate with Okta Verify and a push notification (MacOS, Windows)

```bash
aws sso login --no-browser | headless-aws-sso-with-okta --okta-auth push-notification
```

### Flags
```
-capture-state
    Take screenshots and dump html of each login page.
-email string
    email to sign in with. Okta FastPass will be used if not specified.
-okta-auth string
    Okta authentication method - "fastpass" or "push-notification". (default "push-notification")
-rod string
    Set the default value of options used by rod.
-show-browser
    Show browser window during login.
-verbose
    Print verbose output.
-version
    Print the version and exit.
```