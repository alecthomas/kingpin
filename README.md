# Kingpin - A Go (golang) command line and flag parser [![Build Status](https://travis-ci.org/alecthomas/kingpin.png)](https://travis-ci.org/alecthomas/kingpin)

## Features

- POSIX-style short flag combining.
- Parsed, type-safe flags.
- Parsed, type-safe positional arguments.
- Support for required flags and required positional arguments
- Callbacks per command, flag and argument.

## Simple Example

Kingpin can be used for simple flag+arg applications like so:

```shell
$ ping --help
usage: ping [<flags>] <ip> [<count>]

Flags:
  --debug            Enable debug mode.
  --help             Show help.
  -t, --timeout=5s   Timeout waiting for ping.

Args:
  <ip>        IP address to ping.
  [<count>]   Number of packets to send
$ ping 1.2.3.4 5
Would ping: 1.2.3.4 with timeout 5s%
```

From the following source:

```go
package main

import (
  "fmt"

  "github.com/alecthomas/kingpin"
)

var (
  debug   = kingpin.Flag("debug", "Enable debug mode.").Bool()
  timeout = kingpin.Flag("timeout", "Timeout waiting for ping.").Required().Short('t').Duration()
  ip      = kingpin.Arg("ip", "IP address to ping.").Required().IP()
  moo     = kingpin.Arg("moo", "moo").String()
)

func main() {
  kingpin.Parse()
  fmt.Printf("Would ping: %s with timeout %s", *ip, *timeout)
}
```

## Complex Example

Kingpin can also produce complex command-line applications with global flags,
subcommands, and per-subcommand flags, like this:

```shell
$ chat
usage: chat [<flags>] <command> [<flags>] [<args> ...]

Flags:
  --debug              enable debug mode
  --help               Show help.
  --server=127.0.0.1   server address

Commands:
  help <command>
    Show help for a command.

  post [<flags>] <channel>
    Post a message to a channel.

  register <nick> <name>
    Register a new user.

$ chat help post
usage: chat [<flags>] post [<flags>] <channel> [<text>]

Post a message to a channel.

Flags:
  --image=IMAGE   image to post

Args:
  <channel>   channel to post to
  [<text>]    text to post
$ chat post --image=~/Downloads/owls.jpg pics
...
```

From this code:

```go
package main

import "github.com/alecthomas/kingpin"

var (
  debug    = kingpin.Flag("debug", "enable debug mode").Default("false").Bool()
  serverIP = kingpin.Flag("server", "server address").Default("127.0.0.1").MetaVarFromDefault().IP()

  register     = kingpin.Command("register", "Register a new user.")
  registerNick = register.Arg("nick", "nickname for user").Required().String()
  registerName = register.Arg("name", "name of user").Required().String()

  post        = kingpin.Command("post", "Post a message to a channel.")
  postImage   = post.Flag("image", "image to post").File()
  postChannel = post.Arg("channel", "channel to post to").Required().String()
  postText    = post.Arg("text", "text to post").String()
)

func main() {
  switch kingpin.Parse() {
  // Register user
  case "register":
    println(*registerNick)

  // Post message
  case "post":
    if *postImage != nil {
    }
    if *postText != "" {
    }
  }
}
```

## Parsers

Kingpin supports both flag and positional argument parsers for converting to
Go types. For example, some included parsers are `Int()`, `Float()`,
`Duration()` and `ExistingFile()`.

A goal of Kingpin is to make extending the supported types simple. As an
example, here's the source for the builtin IP parser:

```go
func IP(s Settings) (target *net.IP) {
  target = new(net.IP)
  s.SetParser(func(value string) error {
    if ip := net.ParseIP(value); ip == nil {
      return fmt.Errorf("'%s' is not an IP address", value)
    } else {
      *target = ip
      return nil
    }
  })
  return
}
```

If this weren't a builtin parser you would use it like so:

```go
ip = IP(cmd.Flag("ip", "IP address of server.").Required())
```
