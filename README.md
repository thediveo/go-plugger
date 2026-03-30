# plugger

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/go-plugger/v3)
[![License](https://img.shields.io/github/license/thediveo/go-plugger)](https://img.shields.io/github/license/thediveo/go-plugger)
![build and test](https://github.com/thediveo/go-plugger/actions/workflows/buildandtest.yaml/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/go-plugger/v3)](https://goreportcard.com/report/github.com/thediveo/go-plugger/v3)
![Coverage](https://img.shields.io/badge/Coverage-99.3%25-brightgreen)

`plugger/v3` is a minimalist Go plugin manager featuring type-safe handling of
functions and interfaces (“symbols”) exposed by plugins. Type safety is checked
at compile time, thanks to Go Generics. Plugins usually are realized as packages
exposing certain well-defined functions or interfaces by registering these.
Plugin packages can be statically linked to or dynamically loaded by a Go
application binary.

Applications then can retrieve, for instance, a list of the exposed plugin
functions (“symbols”) of a _specific type_ and then call all of these exposed
plugin functions one after another – and without having to explicitly maintain a
dedicated list of package functions to call in code. As practice shows, such
lists quickly tend to get forgotten when adding new plugins.

`plugger/v3` ensures a well-defined order of the symbols of the same type, where
the symbols are either sorted lexicographically based on plugin names or
optionally using ”placement hints”. This supports such use cases where some of
the plugins might actually build upon the results from plugins that were invoked
earlier.

Another use case is an application retrieving the exposed symbol for only a
particular single named plugin and invoking only this particular plugin.

Finally, `plugger/v3` is safe for concurrent use (as opposed to v0/v2 that are
not).

## Installation

To add `plugger/v3` to your Go module as a dependency:

```bash
go get github.com/thediveo/go-plugger/v3@latest
```

## Usage

Just three steps...

### Define Exposed Symbol Type

First, define a type for the symbol you want to expose by your plugins; this
must be either a function or interface (but not a pure type-constraining
interface). This type will then be used by `plugger/v3` to manage different
exposed symbol types in separate so-called "_groups_".

```go
type pluginFn func() string
```

Define this type only in one place and then import it into your plugins as well
as in the places where you need to work with the exposed symbol(s). Using a
dedicated package just for the exposed symbol type might at first look like
overkill but actually is your friend against import cycles.

### Registering Exposed Symbols

Second, in your plugins, register (expose) the respective `pluginFn`
implementations by fetching the group object for your specific symbol type and
then calling `Register` on it.

```go
import "github.com/thediveo/go-plugger/v3"

func init() {
    plugger.Group[pluginFn]().Register(PluginFn)
}

func PluginFn() string { return "foo" }
```

Please note that `plugger/v3` defaults to deriving the plugin name from the name
of the package from where `Register` is called.

### Calling Exposed Symbols

Finally, when you want to invoke the registered symbols, grab the group object
for your specific symbol type and then range over the group's exposed symbols.

```go
import (
    "github.com/thediveo/go-plugger/v3"
    // ...
    // don't forget to underline-import your (static) plugins!
)

func main() {
    pluginFnGroup := plugger.Group[pluginFn]()
    for _, pluginFn := range pluginFnGroup.Symbols() {
        fmt.Println(pluginFn())
    }
}
```

## Dynamically Loading Plugins

Please see also `example/dynplug` for a working example.

1. make sure your plugin has a `main` package with an empty `main` function.
2. build your plugin shared object using `go build -tags plugger_dynamic
   -buildmode=plugin ...`
   - Please don't forget to specify the `plugger_dynamic` build tag/constraint;
     otherwise, trying to automatically discover and load plugins using
     `dyn.Discover` will panic with a notice to enable the `plugger_dynamic`
     build tag.
3. in you application, call `dyn.Discover` to discover plugins in a specific
   directory (and sub directories) and to load them.

## Migrating from v0/v2 to v3

In `plugger/v3`, groups now correspond with exactly _one_ symbol type, whereas
v0/v2 allowed to register multiple symbols for the same plugin in the same
group. In v3, simply use multiple and now type-safe groups as needed, one for
each type of exposed symbol.

As one benefit, exposed symbols are now inherently nameless from the perspective
of the plugin manager, so no more need to deal with them. And another benefit is
that groups are also nameless too, but instead they are now (symbol) typed.

In v3, exposed symbols are simply registered using their corresponding type-safe
and name-less group, and with the only options available being `WithName(name)`
and `WithPlacement(hint)`.

```go
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
```

## In Unit Tests

Sometimes, unit tests need a well-defined isolated plugin group configuration.
For this, `PluginGroup[T]` objects as returned by `Group[T]()` can now be backed
up and restored using `PluginGroup[T].Backup()` and `PluginGroup[T].Restore()`.
Additionally, `PluginGroup[T].Clear()` resets a plugin group to its initial
empty state.

## DevContainer

> [!CAUTION]
>
> Do **not** use VSCode's "~~Dev Containers: Clone Repository in Container
> Volume~~" command, as it is utterly broken by design, ignoring
> `.devcontainer/devcontainer.json`.

1. `git clone https://github.com/thediveo/go-plugger`
2. in VSCode: Ctrl+Shift+P, "Dev Containers: Open Workspace in Container..."
3. select `plugger.code-workspace` and off you go...

## Supported Go Versions

`lxkns` supports versions of Go that are noted by the [Go release
policy](https://golang.org/doc/devel/release.html#policy), that is, major
versions _N_ and _N_-1 (where _N_ is the current major version).

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Copyright and License

`plugger` is Copyright 2019-2026 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
