/*
Package plugger is a simplistic plugin manager that works with both statically
linked as well as dynamically linked Go plugins. It supports multiple plugin
groups, as well as controlling the order of plugins within a group.

plugger allows a applications to discover the self-registering plugins available
to it and to then invoke their exported plugin functions. plugger supports
organizing plugins into multiple groups, where the plugins from one group
typically feature functionality different from other plugin groups.

# Basic Usage

To work with the plugins registered in the group named "plugins" in order to
call their exported functions, get a group object (of type [PluginGroup]) by
calling [New], specifying the group name:

	plugs := plugger.New("plugins")

There's no need for callers to cache the returned plugin group object; in fact,
it is perfectly fine to simply call [New] whenever working with the plugins of a
group. plugger internally automatically caches the plugin group objects and
always returns the same for a particular named group.

The plugins in a group are ordered to support applications where the plugins
need to be called in specific sequence. Plugin order defaults to lexicographical
order, but can be selectively overridden by the plugins themselves, demanding
specific placement. This supports more sophisticated plugin schemes, where the
plugins allow to modularize and sequence certain application functionality.

	for _, doit := plugs.Func("DoIt") {
	    err := doit.(func() error)()
	    if err != nil {
	        break
	    }
	}

# Static Plugins

Static plugins are permanently linked into the final application, so the
included plugin set cannot be changed after delivering the application binary.

A typical usage pattern is to "dash"-import the static plugin packages in some
convenient place, so the Go toolchain will link them into the final binary and
call their init functions.

	import (
	    _ "example.org/plagueins/delta"
	    _ "example.org/plagueins/omicron"
	)

# Dynamic Plugins

In contrast, dynamically linked plugins are in the form of .so shared library
files which are separate from the application binary. Subject to the usual Go
runtime restrictions for shared Go libraries (for instance, see Eli Bendersky's
[Plugins in Go] blog post and spencer's [go, shared libraries, and ABIs]), the
set of plugins for an application can be changed even after delivery. This also
allows for simplified addition of 3rd party plugins to a plugin-aware
application.

Dynamic plugins can be easily discovered at application runtime from the
filesystem using plugger's dyn package. It supports both flat dynamic plugin
file layout, as well as hierarchical subdirectories, where only the plugin's
root directory is specified.

	dyn.Discover("./pugins", true)

# Plugin Self-Registration

Plugins of applications using the the plugger package must register themselves
in order to become discoverable. In most cases, plugger will be able to
automatically derive the name of a self-registering plugin as well as its group
from a plugin's package name and its parent directory; please see below for
details.

In any case, each plugin needs to register its functions it wants to export and
expose. This is done by calling [Register] and at least the [WithSymbol] option.

	func init() {
	    plugger.Register(plugger.WithSymbol(DoIt))
	}

	func DoIt() {
	    // ...whatever it is you wanna do you plague-in doin'
	}

In some situations, the exported functions of the plugins might need to be run
some specific order. For instance, before or after another particular plugin, or
as early or even as late as possible. Plugger allows plugins to register their
placement demands, such as after another plugin named "foo":

	func init() {
	    plugger.Register(plugger.WithPlacement(">foo"), plugger.WithSymbol(DoIt))
	}

[Plugins in Go]: https://eli.thegreenplace.net/2021/plugins-in-go/
[go, shared libraries, and ABIs]: https://sclem.dev/posts/go-abi/
*/
package plugger
