// Package kingpin provides command line interfaces like this:
//
//    $ chat server <ip>
//    $ chat [--debug] register [--name <name>] <nick>
//    $ chat post --channel|-c <channel> [--image <image>] [<text>]
//
// From code like this:
//
//    var (
//      chat  = kingpin.New("chat", "A command line chat application.")
//      debug = chat.Flag("debug", "enable debug mode").Default("false").Bool()
//
//      server   = chat.Command("server", "Server to connect to.")
//      serverIP = server.Arg("server", "server address").Required().IP()
//
//      register     = chat.Command("register", "Register a new user.")
//      registerName = register.Flag("name", "name of user").Required().String()
//      registerNick = register.Arg("nick", "nickname for user").Required().String()
//
//      post        = chat.Command("post", "Post a message to a channel.")
//      postChannel = post.Flag("channel", "channel to post to").Short('c').Required().String()
//      postImage   = post.Flag("image", "image to post").File()
//      postText    = post.Arg("text", "text to post").String()
//    )
//
//    func main() {
//      switch kingpin.Parse() {
//      case "register":
//        // Register user
//        println(*registerNick)
//
//      case "post":
//        // Post message
//        if *postImage != nil {
//        }
//        if *postText != "" {
//        }
//      }
//    }
package kingpin
