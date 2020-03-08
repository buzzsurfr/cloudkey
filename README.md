# ☁️🔑 cloudkey
Manage your cloud access keys

[![Actions](https://github.com/buzzsurfr/cloudkey/workflows/Go/badge.svg)](https://github.com/buzzsurfr/cloudkey)
[![Join the chat at https://gitter.im/buzzsurfr/cloudkey](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/buzzsurfr/cloudkey?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/buzzsurfr/cloudkey)](https://goreportcard.com/report/github.com/buzzsurfr/cloudkey)
[![GoDoc](https://godoc.org/github.com/buzzsurfr/cloudkey?status.svg)](https://godoc.org/github.com/buzzsurfr/cloudkey)
[![codecov](https://codecov.io/gh/buzzsurfr/cloudkey/branch/master/graph/badge.svg)](https://codecov.io/gh/buzzsurfr/cloudkey)

**Contents**
* [Install](#install)
* [Overview](#overview)
* [Commands](#commands)
  * [`list`](#list)
  * [`rotate`](#rotate)
  * [`version`](#version)

## Install

```console
go install github.com/buzzsurfr/cloudkey
```

## Overview

Cloudkey is a simple way to manage multiple access keys from different cloud providers.

```output
Usage:
  cloudkey [command]

Available Commands:
  help        Help about any command
  list        Lists all cloud access keys
  rotate      Rotate the cloud access key
  version     Version will output the current build information

Global Flags:
      --cloud string    Cloud Provider (default "aws")
      --config string   config file (default is $HOME/.cloudkey.yaml)
      -h, --help            help for cloudkey

Use "cloudkey [command] --help" for more information about a command.
```

## Commands

### `list`

List pulls the credentials from environment variables and the credentials file and outputs them into a table. The "active" profile (which will be rotated by default or used with AWS CLI commands) will be in yellow text.

```output
Usage:
  cloudkey list [flags]

Flags:
  -h, --help            help for list
  -o, --output string   Output format. Only supports 'wide'. (default "table")
```

Example Output:
```output
CLOUD   NAME      ACCESS KEY ID          SOURCE
aws               AKIA************MPLE   EnvironmentVariable
aws     corp      AKIA************CORP   ConfigFile
aws     default   AKIA************G7UP   ConfigFile
aws     lab0      AKIA************FFKG   ConfigFile
aws     lab1      AKIA************YY42   ConfigFile
```

By default, the output type is `table`. You can change the output to `wide` and cloudkey will query AWS to get the account number and UserName associated with each key.

```output
CLOUD   NAME      ACCOUNT        USERNAME       ACCESS KEY ID          SOURCE
aws               123456789012   myUser         AKIA************MPLE   EnvironmentVariable
aws     corp      234567890123   corpUser       AKIA************CORP   ConfigFile
aws     default   012345678901   defaultUser    AKIA************G7UP   ConfigFile
aws     lab0      987654321098   labUser0       AKIA************FFKG   ConfigFile
aws     lab1      987654321098   labUser1       AKIA************YY42   ConfigFile
```

### `rotate`

Rotate uses the "active" access key (or the access key found with the `--profile` option) to request a new access key, applies the access key locally, then uses the new access key to remove the old access key.

Rotate will replace the access key in the same destination as the source, so environment variables are replaced or the config file (credentials file) is modified.

```output
Usage:
  cloudkey rotate [flags]

Flags:
  -h, --help             help for rotate
  -p, --profile string   Profile to rotate
```

### `version`

Version specifies the version, commit, and commit date in either JSON or YAML format.

```output
Version will output the current build information

Usage:
  cloudkey version [flags]

Flags:
  -h, --help            help for version
  -o, --output string   Output format. One of 'yaml' or 'json'. (default "json")
  -s, --short           Print just the version number.
```