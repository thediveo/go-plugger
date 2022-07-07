package fooplug

import plugger "github.com/thediveo/go-plugger/v2"

// PlugFunc is an exported plugin functionality.
func PlugFunc() string { return "fooplug" }

// DoRegister is used for unit testing, as we cannot use the usual automatic
// self-registration using an init() function, but instead have to call it
// explicitly during the flow of unit tests at the right time.
func DoRegister() {
	plugger.Register(plugger.WithPlacement("<barplug"), plugger.WithSymbol(PlugFunc))
}
