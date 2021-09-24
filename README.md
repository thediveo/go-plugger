# plugger

[![Go Reference](https://pkg.go.dev/badge/github.com/thediveo/go-plugger.svg)](https://pkg.go.dev/github.com/thediveo/go-plugger)
![GitHub](https://img.shields.io/github/license/thediveo/go-asciitree)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/go-plugger)](https://goreportcard.com/report/github.com/thediveo/go-plugger)

`plugger` is a simplistic plugin manager that works with both statically linked
as well as dynamically linked Go plugins. It supports multiple plugin groups, as
well as controlled plugin order within a group. Plugins then register named
functions or named interfaces belonging to specific groups. Other application
code finally queries the registered functions and interfaces, and then calls
them as needed.

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
    _ "github.com/thediveo/go-plugger/examples/staticplugins/plugins/foo"
    _ "..."
)

plugs := plugger.New("plugins")
for _, doit := range plugs.Func("DoIt") {
    fmt.Println(doit.(func() string)())
}
```

The plugins get statically linked in by importing them, such as `plugins/foo`.
While at first this might seem to be much overhead, the more plugins you have
in your application, and the more groups you need them to organize into, the
more you'll benefit from the `go-plugger` package: you only need to import
the plugin packages, and plugger will do the rest.

For a more elaborate "example", please also look at `internal/staticplugin/`
and `internal/dynamicplugin/` (these are the built-in test cases).

Please note that in order to use dynamically loaded plugins, the build tag
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
    // import your plugin interface type, say, "I".
    "github.com/.../plugin"
    "github.com/thediveo/go-plugger"
)

type I struct {}
var _ plugin.I = (plugin.I)(nil) // ensure I is actually implemented

func init() {
    RegisterPlugin(&PluginSpec{
        Group: "group",
        Name:  "plug1",
        Symbols: []Symbol{
            NamedSymbol{
                Name:   "I",
                Symbol: plugin.I(&I{}),
            }},
          })
}
```

```go
import (
    // import your plugin interface type, say "I".
    "github.com/.../plugin"

    // import your plugins
    _ "github.com/.../plugin/foo"
    _ "..."
)

plugs := plugger.New("plugins")
for _, i := range plugs.Func("I") {
    fmt.Println(i.(I).DoIt())
}
```

## Run Unit Tests

- VisualStudio Code: please first build the workspace, before running
  tests.

- from CLI: simply run the `./testall.sh` script; it will build the shared
  library for the test shared library plugin(s).

## Copyright and License

`plugger` is Copyright 2019-2021 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
