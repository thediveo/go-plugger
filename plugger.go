// Copyright 2019, 2022 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugger

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// A Symbol is either an exported plugin function or an interface (more precise,
// a pointer to a struct type). Everything else is considered by plugger to be
// an invalid Symbol.
//
// Please that Go's runtime only stores a pointer's
type Symbol interface{}

// NamedSymbol is a [Symbol] with its name explicitly specified, as opposed to
// automatically being derived from the function a Symbol points to.
type NamedSymbol struct {
	Name   string // Name of the Symbol (of the exported plugin function).
	Symbol Symbol // exported plugin function.
}

// PluginSpec describes the registration data of registered plugins returned by
// [PluginGroup.Plugins] and indirectly by [PluginGroup.PluginsFunc]. The
// plugger v2 now avoids registering plugins to have to deal with PluginSpec
// directly, but instead offer a set of registration option functions, such as
// [WithName], [WithGroup], [WithSymbol], et cetera.
//
// Most fields in a PluginSpec can be left zeroed
// (https://golang.org/ref/spec#The_zero_value), with the exception of the
// exported functions (symbols) list.
//
// If left zeroed (that is, unspecified), the plugin name and its group will be
// deduced from the source code path of the plugin .go file that is calling the
// RegisterPlugin function and the PluginSpec updated accordingly.
type PluginSpec struct {
	// optional name. Will be discovered by RegisterPlugin if zero.
	Name string
	// optional plugin group name; corresponds with plugger (group) name. Will
	// be discovered by RegisterPlugin if zero.
	Group string
	// optional placement hint, or ""; placements are in the form "<" (at
	// beginning), "<foo" (before plugin "foo"), ">foo" (after plugin "foo"),
	// and ">" (at end).
	Placement string
	// list of exported plugin functions.
	Symbols []Symbol
	// internal map for quickly resolving symbol names into their interfaces.
	symbolmap map[string]Symbol
}

// PluginFunc is a particular exported named function in a specific plugin. This
// information can be used for logging purposes in applications using plugins,
// such as logging the exact plugin order in which a specific function is called
// on one plugin after another.
type PluginFunc struct {
	F      Symbol
	Name   string
	Plugin *PluginSpec
}

// The internal map of plugin groups being managed.
var pluginGroups = map[string]*PluginGroup{}

// Register registers a plugin by name and group, together with its exported
// plugin functions. Usually, Register is directly called by plugins in their
// init functions, regardless of whether they are statically linked or
// dynamically loaded shared library plugins.
//
// Note: a particular plugin name within a group can be registered only once.
// Any attempt to register the same plugin name twice in the same group will
// result in a panic. This avoids hard-to-diagnose errors which would otherwise
// silently creep in as soon as using placements relative to double-registered
// names. Believe us, we've fallen into the ambigous name trap ourselves.
//
// In the most simple case, a caller to Register only needs to specify an
// exported function to be registered:
//
//	Register(WithSymbol(DoIt))
//
// Further options (see [RegisterOption]) exist to (explicitly) specify
// additional aspects, such as:
//
//   - [WithName] to specify the plugin name instead of deriving it from the
//     package name of the caller.
//   - [WithGroup] to specify the plugin group instead of taking the parent
//     directory name of the plugin's package.
//   - [WithPlacement] to place the plugin in the ordered list of registered
//     plugins at a specific position, such as begin, end, before/after another
//     specific plugin.
//   - [WithSymbol] to register an exported function, the the symbol name
//     taken from the exported function's name.
//   - [WithNamedSymbol] to register an exported function and explicitly setting
//     its symbol name.
//
// For convenience, the plugin name and/or group might be left unspecified
// (zeroed): in this case, Register tries to discover them automatically and
// based on the caller's filename path elements. In order to handle static and
// dynamic (.so) plugins in a uniform manner for registration, only the filename
// information of the caller is used. (Thus, the caller's package name is
// ignored, as in the case of dynamic plugins this would always be "main", and
// we don't want different discovery rules for static and dynamic plugins.)
//
// Given a plugin self-registration function (preferably an init function) in
// file ".../myproj/plugins/foo/plug.go" calls Register, and if no plugin name
// has been specified explicitly, then the plugin name is taken from the name of
// the directory (not path) where the caller is located in. In our case, the
// plugin name would be "foo".
//
// For the same caller, if no (plugin) group name has been specified, then the
// group name is taken from the name of the parent(!) directory of the plugin
// directory. For our case, the group name would be "plugins".
//
// Please note: as always, it is good practice to use the package name also as
// the directory name of the package to avoid confusion; this also allows
// converting static plugins into dynamic plugins rather easily. And
// consequently, you really should better not repeat the mistakes in terms of
// repository and package naming made when creating this module.
func Register(opts ...RegisterOption) {
	registerPlugin(runtime.Caller, opts...)
}

// Aiming for 101% coverage, we go crazy and mock runtime.Caller() ... this
// gives "test-driven development" yet another slightly mad meaning...
func registerPlugin(
	runtimeCaller func(int) (uintptr, string, int, bool),
	opts ...RegisterOption,
) {
	var pspec PluginSpec
	for _, opt := range opts {
		opt(&pspec)
	}
	completePluginSpec(runtimeCaller, &pspec)
	// Discover the names of the exported functions and store them in an index
	// for later lookup. The name discovery can be skipped for exported
	// functions where the caller supplied a NamedSymbol instead.
	pspec.symbolmap = map[string]Symbol{}
	for _, symbol := range pspec.Symbols {
		symname, symbol := resolveSymbol(&pspec, symbol)
		if _, ok := pspec.symbolmap[symname]; ok {
			panic(fmt.Sprintf("Register: plugin %q in group %q, duplicate symbol %q",
				pspec.Name, pspec.Group, symname))
		}
		pspec.symbolmap[symname] = symbol
	}
	// Let's find the plugin's group, or if there isn't one, create a new one
	// for it and other plugins to come that want to share this group.
	pg, ok := pluginGroups[pspec.Group]
	if !ok {
		pg = &PluginGroup{Group: pspec.Group}
		pluginGroups[pspec.Group] = pg
	}
	// Just tack on this plugin to the list of registered plugins in this group.
	// Sorting has to wait for later... Make sure that the same plugin name
	// cannot be registered twice within the same plugin group.
	for _, plug := range pg.plugins {
		if pspec.Name == plug.Name {
			panic(fmt.Sprintf("Register: duplicate plugin name registration '%s'", pspec.Name))
		}
	}
	pg.unordered = true
	pg.plugins = append(pg.plugins, &pspec)
}

// completePluginSpec fills in any missing plugin name and/or group data,
// glancing them from the passed caller information.
func completePluginSpec(runtimeCaller func(int) (uintptr, string, int, bool), pspec *PluginSpec) {
	pathsep := string(os.PathSeparator)
	// Try to fill in "empty" plugin name and group, where necessary.
	if pspec.Name == "" || pspec.Group == "" {
		_, file, _, ok := runtimeCaller(3)
		if !ok {
			panic("Register: unable to discover caller for discovering name and/or group plugin information")
		}
		plugdir := filepath.Dir(file)
		if pspec.Name == "" {
			pspec.Name = filepath.Base(plugdir)
		}
		if pspec.Group == "" {
			pspec.Group = filepath.Base(filepath.Dir(plugdir))
		}
	}
	if pspec.Name == "" || pspec.Name == "." || pspec.Name == pathsep ||
		pspec.Group == "" || pspec.Group == "." || pspec.Group == pathsep {
		panic("Register: missing plugin name and/or group registration information")
	}
}

// resolveSymbol checks a given Symbol to be either a function or [NamedSymbol],
// determining its name if necessary, and returing the symbol's name and
// function.
func resolveSymbol(pspec *PluginSpec, symbol Symbol) (string, Symbol) {
	if namedsym, ok := symbol.(NamedSymbol); ok {
		if namedsym.Name == "" {
			panic(fmt.Sprintf("Register: plugin %q in group %q, zero-named symbol %#v",
				pspec.Name, pspec.Group, symbol))
		}
		switch reflect.TypeOf(namedsym.Symbol).Kind() {
		case reflect.Func, reflect.Ptr:
			break
		default:
			panic(fmt.Sprintf("Register: plugin %q in group %q, invalid symbol %#v",
				pspec.Name, pspec.Group, symbol))
		}
		return namedsym.Name, namedsym.Symbol
	}
	var symname string
	t := reflect.TypeOf(symbol)
	switch t.Kind() {
	case reflect.Func:
		symname = strings.SplitN(
			filepath.Base(runtime.FuncForPC(
				reflect.ValueOf(symbol).Pointer()).Name()),
			".", 2)[1]
	case reflect.Ptr:
		e := t.Elem()
		if e.Kind() != reflect.Struct {
			panic(fmt.Sprintf("Register: plugin %q in group %q, invalid symbol type %t",
				pspec.Name, pspec.Group, symbol))
		}
		symname = t.Elem().Name()
	default:
		panic(fmt.Sprintf("plugin %q in group %q: invalid symbol type %t",
			pspec.Name, pspec.Group, symbol))
	}
	return symname, symbol
}
