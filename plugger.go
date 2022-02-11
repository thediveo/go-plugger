// Copyright 2019 Harald Albrecht.
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
	"sort"
	"strings"
)

// A Symbol is a pointer to an exported plugin function.
type Symbol interface{}

// NamedSymbol is a Symbol where its name is explicitly specified, as opposed to
// automatically being derived from the function a Symbol points to.
type NamedSymbol struct {
	Name   string // Name of the Symbol (of the exported plugin function).
	Symbol Symbol // exported plugin function.
}

// PluginSpec describes the registration data of a plugin. Most of its fields
// can be left zeroed (https://golang.org/ref/spec#The_zero_value), with the
// exception of the exported functions (symbols) list. This list of exported
// functions must not be empty, otherwise the plugin registration will be
// ignored.
//
// If left zeroed (that is, unspecified), the plugin name and its group will be
// deduced from the source code path of the plugin .go file that is calling the
// RegisterPlugin function.
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

// PluginGroup is a named group of plugins that lists those plugins registered
// for this group in "sorted" order. Plugins that do not register any placement
// demands are sorted in lexicographic order.
type PluginGroup struct {
	Group    string        // group of plugins this plugger manages.
	unsorted bool          // has the list of registered plugins been sorted or is it dirty?
	plugins  []*PluginSpec // sorted list of registered plugins (plugin specifications).
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

// RegisterPlugin registers a plugin by name and group with its exported plugin
// functions. This registration function is to be called by plugins, regardless
// of whether they are statically linked or dynamically loaded shared library
// plugins.
//
// Note: the same plugin name within a group can be registered only once. Any
// attempt to register the same plugin name twice in the same group will result
// in a panic. This avoids hard-to-diagnose errors which would otherwise
// silently creep in as soon as using placements relative to double-registered
// names.
//
// In addition to its exported plugin functions, a plugin might also specify its
// placement within its plugin group: at the beginning, end, or before/after
// another (named) plugin within the same group.
//
//   - Placement: "<" ... at beginning of plugin list
//   - Placement: "<foo" ... just before plugin "foo"
//   - Placement: ">" ... at end of plugin list
//   - Placement: ">foo" ... directly after plugin "foo"
//
// For convenience, the plugin name and/or group might be left unspecified
// (zeroed): in this case, RegisterPlugin tries to discover them automatically
// and based on the caller's filename path elements. In order to handle static
// and dynamic (.so) plugins in a uniform manner for registration, only the
// filename information of the caller is used. (Thus, the caller's package name
// is ignored, as in the case of dynamic plugins this would always be "main",
// and we don't want different discovery rules for static and dynamic plugins.)
//
// Given a plugin function (preferably an init function) in file
// ".../myproj/plugins/foo/plug.go" calls RegisterPlugin, and if no plugin name
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
// converting static plugins into dynamic plugins rather easily.
func RegisterPlugin(plugspec *PluginSpec) {
	registerPlugin(plugspec, runtime.Caller)
}

// Aiming for 101% coverage, we go crazy and mock runtime.Caller() ... this
// gives "test-driven development" yet another slightly mad meaning...
func registerPlugin(plugspec *PluginSpec,
	runtimeCaller func(int) (uintptr, string, int, bool)) {
	pathsep := string(os.PathSeparator)
	// Work on a copy of the plugin specification as we might need to fill in
	// some fields that the caller did not specify.
	pspec := *plugspec
	// Try to fill in "empty" plugin name and group, where necessary.
	if pspec.Name == "" || pspec.Group == "" {
		_, file, _, ok := runtimeCaller(1 + 1) // NOT a spelling mistake!
		if !ok {
			panic("unable to discover caller for discovering name and/or group plugin information")
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
		panic("missing plugin name and/or group registration information")
	}
	// Discover the names of the exported functions and store them in an index
	// for later lookup. The name discovery can be skipped for exported
	// functions where the caller supplied a NamedSymbol instead.
	pspec.symbolmap = map[string]Symbol{}
	for _, symbol := range pspec.Symbols {
		if namedsym, ok := symbol.(NamedSymbol); ok {
			if namedsym.Name == "" {
				continue
			}
			switch reflect.TypeOf(namedsym.Symbol).Kind() {
			case reflect.Func, reflect.Ptr:
				break
			default:
				continue
			}
			if _, ok = pspec.symbolmap[namedsym.Name]; ok {
				panic(fmt.Sprintf("plugin %q in group %q: duplicate symbol %q",
					pspec.Name, pspec.Group, namedsym.Name))
			}
			pspec.symbolmap[namedsym.Name] = namedsym.Symbol
		} else {
			var symname string
			t := reflect.TypeOf(symbol)
			switch t.Kind() {
			case reflect.Func:
				symname = strings.SplitN(filepath.Base(runtime.FuncForPC(
					reflect.ValueOf(symbol).Pointer()).Name()), ".", 2)[1]
			case reflect.Ptr:
				symname = t.Elem().Name()
			default:
				continue
			}
			if _, ok = pspec.symbolmap[symname]; ok {
				panic(fmt.Sprintf("plugin %q in group %q: duplicate symbol %q",
					pspec.Name, pspec.Group, symname))
			}
			pspec.symbolmap[symname] = symbol
		}
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
			panic(fmt.Sprintf("duplicate plugin name registration '%s'", pspec.Name))
		}
	}
	pg.unsorted = true
	pg.plugins = append(pg.plugins, &pspec)
}

// New returns the plugin group object of a given name, or an empty group object
// in case there never registered any plugins in the group specified by the
// caller. The plugin group object then provides access to the plugins' exported
// functions in order to call plugin functionality.
func New(name string) *PluginGroup {
	pg, ok := pluginGroups[name]
	if ok {
		if pg.unsorted {
			pg.sort()
			pg.unsorted = false
		}
	} else {
		pg = &PluginGroup{Group: name}
	}
	return pg
}

// Func returns a slice of interface{}s for all those plugins in a group
// providing the function with "name".
func (pg *PluginGroup) Func(name string) []Symbol {
	fs := make([]Symbol, 0, len(pg.plugins))
	for _, plug := range pg.plugins {
		// Only add those exported symbols that are actually functions.
		f, ok := plug.symbolmap[name]
		if ok && reflect.TypeOf(f).Kind() == reflect.Func {
			fs = append(fs, f)
		}
	}
	return fs
}

// FuncPrefix returns a slice of interface{}s for all plugins in a group
// providing functions which start with "prefix"-
func (pg *PluginGroup) FuncPrefix(prefix string) []Symbol {
	fs := make([]Symbol, 0, len(pg.plugins))
	for _, plug := range pg.plugins {
		// Only add those exported symbols that are actually functions.
		for name, f := range plug.symbolmap {
			if strings.HasPrefix(name, prefix) && reflect.TypeOf(f).Kind() == reflect.Func {
				fs = append(fs, f)
			}
		}
	}
	return fs
}

// PluginsFunc returns a slice with the interface{}s of a specifically named
// exported plugin function, together with the plugins exporting them. This
// information can be useful for logging in applications which specific plugins
// actually get invoked for a certain function.
func (pg *PluginGroup) PluginsFunc(name string) []PluginFunc {
	pf := make([]PluginFunc, 0, len(pg.plugins))
	for _, plug := range pg.plugins {
		f, ok := plug.symbolmap[name]
		if !ok {
			continue
		}
		switch reflect.TypeOf(f).Kind() {
		case reflect.Func, reflect.Ptr:
			pf = append(pf, PluginFunc{
				F:      f,
				Name:   name,
				Plugin: plug,
			})
		}
	}
	return pf
}

// PluginFunc returns the interface{} to the named exported plugin function in
// the named plugin.
func (pg *PluginGroup) PluginFunc(plugin string, name string) Symbol {
	for _, plug := range pg.plugins {
		if plug.Name == plugin {
			f, ok := plug.symbolmap[name]
			if ok && reflect.TypeOf(f).Kind() == reflect.Func {
				return f
			}
			return nil
		}
	}
	return nil
}

// Plugins returns the list of plugin specifications in this group. For
// instance, this can be used to log the actual set of plugins found or
// available.
func (pg *PluginGroup) Plugins() []*PluginSpec {
	return pg.plugins
}

// PluginNames returns the list of plugin names in this group. This is mostly a
// convenience function for logging, unit testing, et cetera.
func (pg *PluginGroup) PluginNames() []string {
	names := make([]string, 0, len(pg.plugins))
	for _, plugin := range pg.plugins {
		names = append(names, plugin.Name)
	}
	return names
}

// Sorts the plugins by name and optionally by reference; that is, individual
// plugins can claim to get to the front/end, or before/after a another named
// plugin. This is with a nod to Jeremy Ruston and his incredible TiddlyWiki
// (and its list and module sorting).
func (pg *PluginGroup) sort() {
	// First, sort lexicographically by plugin name (not: by plugin path).
	sort.Slice(pg.plugins, func(a, b int) bool {
		return pg.plugins[a].Name < pg.plugins[b].Name
	})
	// Second, honor the optional positional requests of individual plugins.
	// Or, at least try to do so...
	plugs := make([]*PluginSpec, len(pg.plugins))
	copy(plugs, pg.plugins)
	for _, plug := range pg.plugins {
		// Find the next plugin to process from the original list on in the
		// current and potentially modified list, because we need to work on the
		// current list when shuffling plugins around.
		var idx int
		var pl *PluginSpec
		for idx, pl = range plugs {
			if pl.Name == plug.Name {
				break
			}
		}
		pos := idx // start with no change in a plugin's sequence position
		// Does the plugin want to be positioned either before a specifically
		// named other plugin or at the beginning?
		if strings.HasPrefix(plug.Placement, "<") {
			before := plug.Placement[1:]
			if before == "" {
				pos = 0 // tangarines FIRST (*all* of them, *snicker*)
			} else {
				// Find the named plugin at its current position; not at the
				// original position, that wouldn't make sense and mix up the
				// original intention.
				for i, p := range plugs {
					if before == p.Name {
						pos = i
						break
					}
				}
			}
		}
		// Does the plugin want to be positioned either after another
		// specifically named plugin or at the end of the sequence?
		if strings.HasPrefix(plug.Placement, ">") {
			after := plug.Placement[1:]
			if after == "" {
				pos = len(plugs)
			} else {
				// Find the named plugin at its current position; not at the
				// original position, that wouldn't make sense and mix up the
				// original intention.
				for i, p := range plugs {
					if after == p.Name {
						pos = i + 1
						break
					}
				}
			}
		}
		// I severely miss Python's and Javascript's simplistic way to move
		// elements within slices. Go is just ugly and terrible. Any of its
		// claims to have been inspired by Python is like Steve Balmer
		// claiming to be inspired by Unix...
		if idx < pos {
			// before: [.] [.] [X] [:] [:] [P] [.]
			// after:  [.] [.] [:] [:] [P] [X] [.]

			// border case: after end
			// before: [.] [.] [X] [:] [:] P
			// after:  [.] [.] [:] [:] [X]

			// border case: at end
			// before: [.] [.] [X] P
			// after:  [.] [.] [P]
			pos--
			for i := idx; i < pos; i++ {
				plugs[i] = plugs[i+1]
			}
			plugs[pos] = plug
		} else if idx > pos {
			// before: [.] [.] [P] [:] [:] [X] [.]
			// after:  [.] [.] [X] [P] [:] [:] [.]

			// before: [P] [:] [:] [X] [.] [.] [.]
			// after:  [X] [P] [:] [:] [.] [.] [.]
			for i := idx; i > pos; i-- {
				plugs[i] = plugs[i-1]
			}
			plugs[pos] = plug
		}
	}
	pg.plugins = plugs
}
