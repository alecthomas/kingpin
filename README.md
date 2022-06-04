# Overview

Fisk is a [fluent-style](http://en.wikipedia.org/wiki/Fluent_interface), type-safe command-line parser. It supports flags, nested commands, and positional arguments.

This is a fork of [kingpin](https://github.com/alecthomas/kingpin), a very nice CLI framework that has been in limbo for a few years. As this project and others we work on are heavily invested in Kingpin we thought to revive it for our needs.

For full help and intro see [kingpin](https://github.com/alecthomas/kingpin), this README will focus on our local changes.

## Versions

We are not continuing the versioning scheme of Kingpin, the Go community has introduced onerous SemVer restrictions, we will start from 0.0.1 and probably never pass 0.x.x.

Some historical points in time are kept:

| Tag    | Description                                                     |
|--------|-----------------------------------------------------------------|
| v0.0.1 | Corresponds to the v2.2.6 release of Kingpin                    |
| v0.0.2 | Corresponds with the master of Kingpin at the time of this fork |
| v0.1.0 | The first release under `choria-io` org                         |

## Notable Changes

 * Renamed `master` branch to `main`
 * Incorporate `github.com/alecthomas/units` and `github.com/alecthomas/template` as local packages
 * Changes to make `staticcheck` happy
