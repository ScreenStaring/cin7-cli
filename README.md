# Cin7 CLI

Command-line program to interact with Cin7.

## Installation

Download the version for your platform on the [releases page](https://github.com/ScreenStaring/cin7-cli/releases).
Windows, macOS/OS X, and GNU/Linux are supported.

### Using [Homebrew](https://brew.sh/)

```
brew tap screenstaring/tap
brew install cin7-cli
```

## Usage

The CLI interface uses the executable `cin7`. Not everything exposed by the Cin7 API is supported:

    NAME:
       cin7 - Utilities for working with Cin7

    USAGE:
       cin7 [global options] [command [command options]]

    VERSION:
       0.0.1

    COMMANDS:
       location, locations, loc     Location utilities
       product, products, p         Product utilities
       stock                        Stock utilities
       webhook, webhooks, hooks, w  Webhook utilities
       help, h                      Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --help, -h     show help
       --version, -v  print the version

## Credentials

To run commands you'll must specify account and key:

```
cin7 COMMAND -a ACCOUNT -k KEY
```

You can set these via [Environment Variables](#environment-variables).

### Environment Variables

You can use the following environment variables to set credentials:

- `CIN7_ACCOUNT_ID`
- `CIN7_APPLICATION_KEY`

## Commands

### Locations

    NAME:
       cin7 location - Location utilities

    USAGE:
       cin7 location [command [command options]]

    COMMANDS:
       list, ls  List locations

    OPTIONS:
       --help, -h  show help

### Products

    NAME:
       cin7 product - Product utilities

    USAGE:
       cin7 product [command [command options]]

    COMMANDS:
       list, ls  List products

    OPTIONS:
       --help, -h  show help


### Stock

    NAME:
       cin7 stock - Stock utilities

    USAGE:
       cin7 stock [command [command options]]

    COMMANDS:
       adjustments, a, adj  Stock adjustment commands
       take, t              Stock take commands

    OPTIONS:
       --help, -h  show help

### Webhooks

    NAME:
       cin7 webhook - Webhook utilities

    USAGE:
       cin7 webhook [command [command options]]

    COMMANDS:
       list, ls            List webhooks for the account
       delete, del, rm, d  Delete the webhook with the given ID
       create, c           Create a webhook
       update, u           Update a webhook

    OPTIONS:
       --help, -h  show help

Note that when updating webhooks Cin7 will not allow some values to be set to `null` and the property will not be updated.

## License

Released under the MIT License: http://www.opensource.org/licenses/MIT

---

Made by [ScreenStaring](https://screenstaring.com)
