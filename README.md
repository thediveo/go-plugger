# plugger

[![Go Reference](https://pkg.go.dev/badge/github.com/thediveo/go-plugger.svg)](https://pkg.go.dev/github.com/thediveo/go-plugger/v3)
![GitHub](https://img.shields.io/github/license/thediveo/go-plugger)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/go-plugger/v3)](https://goreportcard.com/report/github.com/thediveo/go-plugger/v3)
![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)

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
interface). This type will then be used `plugger` to manage different exposed
symbol types in separate so-called "groups".

```go
type pluginFn func() string
```

Define this type only in one place and then import it into your plugins as well
as in the places where you need to work with the exposed symbol(s). Using a
dedicated package just for the exposed symbol type might at first look like
overkill but is your friend against import cycles.

### Registering Exposed Symbols

Second, in your plugins, register (expose) the respective `pluginFn`
implementations by fetching the group object for your specific symbol type and
then calling `Register` on it.

```go
import "github.com/thediveo/go-plugger/v3"

func init() {
    plugger.Group[pluginFn]().Register(MyPluginFn)
}

func MyPluginFn() string { return "foo" }
```

Please note that `plugger/v3` defaults to deriving the plugin name from the
package name where `Register` is called.

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

## VSCode Tasks

The included `go-plugger.code-workspace` defines the following tasks:

- **View Go module documentation** task: installs `pkgsite`, if not done already
  so, then starts `pkgsite` and opens VSCode's integrated ("simple") browser to
  show the go-plugger/v2 documentation.

- **Build workspace** task: builds all, including the shared library test
  plugin.

- **Run all tests with coverage** task: does what it says on the tin and runs
  all tests with coverage.

#### Aux Tasks

- _pksite service_: auxilliary task to run `pkgsite` as a background service
  using `scripts/pkgsite.sh`. The script leverages browser-sync and nodemon to
  hot reload the Go module documentation on changes; many thanks to @mdaverde's
  [_Build your Golang package docs
  locally_](https://mdaverde.com/posts/golang-local-docs) for paving the way.
  `scripts/pkgsite.sh` adds automatic installation of `pkgsite`, as well as the
  `browser-sync` and `nodemon` npm packages for the local user.
- _view pkgsite_: auxilliary task to open the VSCode-integrated "simple" browser
  and pass it the local URL to open in order to show the module documentation
  rendered by `pkgsite`. This requires a detour via a task input with ID
  "_pkgsite_".

## Make Targets

- `make`: lists all targets.
- `make coverage`: runs all tests with coverage and then **updates the coverage
  badge in `README.md`**.
- `make pkgsite`: installs [`x/pkgsite`](golang.org/x/pkgsite/cmd/pkgsite), as
  well as the [`browser-sync`](https://www.npmjs.com/package/browser-sync) and
  [`nodemon`](https://www.npmjs.com/package/nodemon) npm packages first, if not
  already done so. Then runs the `pkgsite` and hot reloads it whenever the
  documentation changes.
- `make report`: installs
  [`@gojp/goreportcard`](https://github.com/gojp/goreportcard) if not yet done
  so and then runs it on the code base.
- `make test`: runs **all** tests (including dynamic plugins).

## Copyright and License

`plugger` is Copyright 2019-2022 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
