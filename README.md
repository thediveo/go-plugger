# plugger

[![Go Reference](https://pkg.go.dev/badge/github.com/thediveo/go-plugger.svg)](https://pkg.go.dev/github.com/thediveo/go-plugger/v2)
![GitHub](https://img.shields.io/github/license/thediveo/go-plugger)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/go-plugger/v2)](https://goreportcard.com/report/github.com/thediveo/go-plugger/v2)
![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)

`plugger` is a simplistic plugin manager that works with both statically linked
as well as dynamically linked Go plugins. It supports multiple plugin groups, as
well as controlled plugin order within a group. Plugins then register named
functions or named interfaces belonging to specific groups. Other application
code finally queries the registered functions and interfaces, and then calls
them as needed.

## Installation

To add `plugger/v2` to your Go module as a dependency (see below for migrating
to v2):

```bash
go get github.com/thediveo/go-plugger/v2@latest
```

## Examples

The plugin registration mechanism supports registering and working with symbols
that are either functions or interface pointers.

### Registering and Calling Functions

See the `examples` directory for how to use plugger in a Go application for
organizing and using static plugins (plugins that have been statically linked
into your application). `examples/staticplugins/main.go` uses plugger to
get all plugins in the "plugin" group and then calls some method on them.

```go
import (
    // import your plugins
    _ "github.com/thediveo/go-plugger/v2/examples/staticplugins/plugins/foo"
    // ...
)

func main() {
    plugs := plugger.New("plugins")
    for _, doit := range plugs.Func("DoIt") {
        fmt.Println(doit.(func() string)())
    }
}
```

The plugins get statically linked in by importing them, such as `plugins/foo`.
While at first this might seem to be much overhead, the more plugins you have
in your application, and the more groups you need them to organize into, the
more you'll benefit from the `go-plugger` package: you only need to import
the plugin packages, and plugger will do the rest.

For a more elaborate "example", please also look at `internal/staticplugin/`
and `internal/dynamicplugin/` (these are the built-in test cases).

Please note that in order to use dynamically loaded plugins, the **build tag**
`plugger_dynamic` needs to be set. The `plugger` module now defaults to **not
including** support for dynamically loading plugins, unless explicitly requested
by the `plugger_dynamic` build tag. The default avoids linker warnings when
building fully static binaries without any dynamic C library references.

### Registering and Calling Interfaces

When registering interfaces it usually will be necessary to explicitly specify
the interface name for registration as otherwise Go's reflection mechanism will
cause the symbol detection to use the name of the implementing struct type
instead. Depending on your coding style that might work, or might not.

```go
import (
    "github.com/thediveo/go-plugger"
    ".../myplugins" // import your plugin interface type, say, "I".
)

type I struct {}
var _ myplugins.I = (*I)(nil) // ensure I implements plugin.I

func init() {
    plugger.RegisterPlugin(plugger.WithName("plug1"),
        plugger.WithGroup("group"),
        plugger.WithNamedSymbol("I", myplugins.I(&I{}))
}
```

```go
import (
    "github.com/thediveo/go-plugger"
    ".../myplugins" // import your plugin interface type, say, "I".
    // import your plugins
    _ ".../myplugins/foo"
    _ "..."
)

plugs := plugger.New("plugins")
for _, i := range plugs.Func("I") {
    fmt.Println(i.(myplugins.I).DoIt())
}
```

## Migrating from v0 to v2

The registration is now done by calling `Register()` (formerly
~~`RegisterPlugin`~~) and takes _option functions_ instead of the unwieldly
`PluginSpec`.

```go
// v0:
//   plugger.RegisterPlugin(&plugger.PluginSpec{
//      Group:   "group",
//      Name:    "plug1",
//      Symbols: []plugger.Symbol{foo},
//   })
//
// v2:
plugger.Register(plugger.WithName("plug1"), 
    plugger.WithGroup("group"), plugger.WithSymbol(foo))
```

The other parts of the `plugger` API remain unchanged in v2, such as
`New(groupname)`, et cetera.

Please note that v2 now is very strict in what gets registered and `panic`s with
details in order to clearly mark registration errors, such as
non-function/interface symbols, duplicate names, …

## Run Unit Tests

- VisualStudio Code: please first build the workspace, before running
  tests related to dynamic plugins.

- from CLI: simply run `make test`.

## Copyright and License

`plugger` is Copyright 2019-2022 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
