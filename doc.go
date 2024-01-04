/*
Package plugger v3 implements a minimalist plugin manager featuring type-safe
handling of functions and interfaces (“symbols”) exposed by plugins. Type safety
is checked at compile time, thanks to Go Generics. Plugins usually are realized
as packages exposing certain well-defined functions or interfaces by registering
these. Plugin packages can be statically linked to or dynamically loaded by a Go
application binary.

Applications then can retrieve, for instance, a list of the exposed plugin
functions (“symbols”) of a specific type using [Group][T]().Symbols() and then
call all of these exposed plugin functions one after another – and without
having to explicitly maintain a dedicated list of package functions to call in
code. As practice shows, such lists quickly tend to get forgotten when adding
new plugins.

Plugger v3 ensures a well-defined order of the symbols of the same type, where
the symbols are either sorted lexicographically based on plugin names or
optionally using ”placement hints”. This supports such use cases where some of
the plugins might actually build upon the results from plugins that were invoked
earlier.

Plugger v3 is safe for concurrent use (as opposed to v0/v2 that are not).

# Usage

Exposed plugin symbols are organized in [PluginGroup] objects, based on their
particular type. The first step thus is to define a dedicated type for an
exposed plugin symbol, such as a function or interface:

	type PluginFn func(string) string

A good practice is to define the exported symbol types in a dedicated and
otherwise empty package. This not only avoids import cycles but also ensures
that always the same symbol type is used for looking up the corresponding
[PluginGroup] object when working with symbols.

The [PluginGroup] for a specific type is retrieved by calling [Group] for the
specific type:

	group := plugger.Group[PluginFn]()

Calling [Group] multiple times for the same type always returns the same
[PluginGroup] instance. There's no need for global variables referencing plugin
group objects and using them should be avoided.

Next, plugins register their exposed symbols by retrieving the symbol's group
first and then calling the [plugger.PluginGroup.Register] receiver on the group
object.

	func init() {
	    plugger.Group[pluginFn]().Register(MyPluginFn)
	}

	func MyPluginFn() string { return "foo" }

Please note that plugger defaults to deriving the plugin name from the package
name where [plugger.PluginGroup.Register] is called. The plugin name can also be
explicitly specifyed by using [WithPlugin] in a registration.

Finally, when an application wants to invoke the registered symbols, it needs to
grab the group object for the specific symbol type as before and then range over
the group's exposed [plugger.PluginGroup.Symbols].

	import (
	    // don't forget to underline-import your (static) plugins!
	)

	func main() {
	    pluginFnGroup := plugger.Group[pluginFn]()
	    for _, pluginFn := range pluginFnGroup.Symbols() {
	        fmt.Println(pluginFn())
	    }
	}

# Dynamically loading Plugins

Specify the build tag/constraint “plugger_dynamic” and use
[github.com/thediveo/go-plugger/v3/dyn.Discover] to discover and load plugin
shared objects.

# Upgrading from v0/v2

Plugger v3 simplifies the API while at the same time introducing type-safety for
the exposed symbols. In v3, a given [PluginGroup] always contains only symbols
of the same particular type, but never multiple different symbol types. In
consequence, the overhead of naming exposed symbols in order to differentiate
them could be removed; this v1/v2 feature wasn't really used anyway.

	// v3:
	plugger.Group[fooFn]().Register(foo)
	// before, v0:
	//   plugger.RegisterPlugin(&plugger.PluginSpec{
	//      Group:   "group",
	//      Name:    "plug1",
	//      Symbols: []plugger.Symbol{foo},
	//   })
	// before, v2:
	// plugger.Register(plugger.WithName("plug1"),
	//     plugger.WithGroup("group"), plugger.WithSymbol(foo))

# In Unit Tests

Sometimes, unit tests need a well-defined isolated plugin group configuration.
For this, [PluginGroup] objects returned by [Group]() can now be backed up and
restored using [PluginGroup.Backup] and [PluginGroup.Restore]. Additionally,
[PluginGroup.Clear] resets a plugin group to its initial empty state.
*/
package plugger
