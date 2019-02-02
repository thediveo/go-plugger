/*
Package plugger is a simplistic plugin manager that works with both statically
linked as well as dynamically linked Go plugins. It supports multiple plugin
groups, as well as controlled plugin order within a group.

It allows an application to discover the plugins available to it and to then
invoke the plugin functions. Plugger supports plugin groups in order to back
multiple groups of plugins, where each plugins from one group feature
different functionality from other plugin groups.

    plugs := plugger.New("plugins")

The plugins in a group are ordered to support applications where the plugins
need to be called in specific sequence. Plugin order defaults to
lexicographical order, but can be selectively overriden by the plugins
themselves, demanding specific placement. This supports more sophisticated
plugin schemes, where the plugins allow to modularize and sequence certain
application functionality.

    for _, doit := plugs.Func("DoIt") {
        done := doit.(func() bool)()
        if done { break }
    }

Static and Dynamic Plugins

Static plugins are permanently linked into the final application, so the
included plugin set cannot be changed afterwards delivery of the application.
In contrast, dynamically linked plugins are in the form of .so shared library
files which are separate from the application binary. Subject to the usual Go
runtime restrictions for shared Go libraries, the set of plugins for an
application can be changed even after delivery. This also allows for
simplified addition of 3rd party plugins to a plugin-aware application.

Dynamic plugins can be easily discovered at application runtime from the
filesystem using plugger's dyn package. It supports both flat dynamic plugin
file layout, as well as hierarchical subdirectories, where only the plugin's
root directory is specified.

    dyn.Discover("./pugins", true)

Plugin Self-Registration

Due to the current design of the Go runtime the plugins of applications using
the the plugger package must automatically register themselves. In most
usecases, plugger will be able to automatically guess the name of a
self-registering plugin as well as its group. For details, please see below.

In any case, a plugin needs to specify the plugin functions it wants to export
and thus to expose.

    func init() {
        plugger.RegisterPlugin(&plugger.PluginSpec{
            Symbols: []plugger.Symbol{DoIt},
        })
    }

    func DoIt() {}

In some situations, some plugins might need to be run some specific order,
such before or after some specific other plugin, or as early or late as
possible. Plugger allows plugins to register their placement demands, such as
after another plugin named "foo":

    func init() {
        plugger.RegisterPlugin(&plugger.PluginSpec{
            Placement: ">foo",
            Symbols: []plugger.Symbol{DoIt},
        })
    }

*/
package plugger
