# Kingpin - A Go (golang) command line and flag parser [![Build Status](https://travis-ci.org/alecthomas/kingpin.png)](https://travis-ci.org/alecthomas/kingpin)

## Features

- POSIX-style short flag combining.
- Parsed, type-safe flags.
- Parsed, type-safe positional arguments.
- Callbacks per command, flag and argument callbacks.

## Example

Kingpin supports command line interfaces like this:


```bash
$ chat server <ip>
$ chat [--debug] register [--name <name>] <nick>
$ chat post --channel|-c <channel> [--image <image>] [<text>]
```

From code like this:

```go
var (
  chat  = kingpin.New("chat", "A command line chat application.")
  debug = chat.Flag("debug", "enable debug mode").Default("false").Bool()

  server   = chat.Command("server", "Server to connect to.")
  serverIP = server.Arg("server", "server address").Required().IP()

  register     = chat.Command("register", "Register a new user.")
  registerName = register.Flag("name", "name of user").Required().String()
  registerNick = register.Arg("nick", "nickname for user").Required().String()

  post        = chat.Command("post", "Post a message to a channel.")
  postChannel = post.Flag("channel", "channel to post to").Short('c').Required().String()
  postImage   = post.Flag("image", "image to post").File()
  postText    = post.Arg("text", "text to post").String()
)

func main() {
  switch kingpin.Parse() {
  case "register":
    // Register user
    println(*registerNick)

  case "post":
    // Post message
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
