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
	"reflect"
	"sort"
	"strings"
)

// PluginGroup is a named group of ordered registered plugins. Plugins that do
// not register any specific placement requirements (such as first, last, or
// before/after another specific plugin) are placed in lexicographic order.
type PluginGroup struct {
	Group     string        // name of group of plugins this plugger manages.
	unordered bool          // has the list of registered plugins been ordered or is it still arbitrary?
	plugins   []*PluginSpec // ordered list of registered plugins (plugin specifications).
}

// New returns the plugin group object of a given group name, or an empty group
// object in case there were never any plugins registered in the specified
// group. The plugin group object then provides access to the registered
// exported functions of plugins in order to call plugin functionality. Multiple
// calls for the same group always return the same PluginGroup object.
func New(name string) *PluginGroup {
	pg, ok := pluginGroups[name]
	if ok {
		if pg.unordered {
			pg.sort()
			pg.unordered = false
		}
	} else {
		pg = &PluginGroup{Group: name}
		pluginGroups[name] = pg
	}
	return pg
}

// Func returns a slice of Symbols for all those plugins in a group actually
// providing the function (symbol) with specified name.
func (pg *PluginGroup) Func(name string) []Symbol {
	fs := make([]Symbol, 0, len(pg.plugins))
	for _, plug := range pg.plugins {
		// Only add registered exported symbols that actually are functions.
		if fn, ok := plug.symbolmap[name]; ok && IsFunc(fn) {
			fs = append(fs, fn)
		}
	}
	return fs
}

// FuncPrefix returns a slice of Symbols for all plugins in a group providing
// functions with names that start with the specified "prefix".
func (pg *PluginGroup) FuncPrefix(prefix string) []Symbol {
	fs := make([]Symbol, 0, len(pg.plugins))
	for _, plug := range pg.plugins {
		// Only add those exported symbols that actually are functions.
		for name, f := range plug.symbolmap {
			if strings.HasPrefix(name, prefix) && IsFunc(f) {
				fs = append(fs, f)
			}
		}
	}
	return fs
}

// PluginsFunc returns a slice with the PluginFuncs of a specifically named
// exported plugin function, together with the plugins exporting them. This
// information can be used in applications to log which concrete plugins get
// invoked for a certain function.
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

// PluginFunc returns the Symbol to the named exported plugin function in the
// named plugin.
func (pg *PluginGroup) PluginFunc(plugin string, name string) Symbol {
	for _, plug := range pg.plugins {
		if plug.Name == plugin {
			f, ok := plug.symbolmap[name]
			if ok && IsFunc(f) {
				return f
			}
			return nil
		}
	}
	return nil
}

// Plugins returns (a copy of) the list of plugin specifications in this group.
// For instance, this can be used to log the actual set of registered plugins.
func (pg *PluginGroup) Plugins() []*PluginSpec {
	return append(pg.plugins[:0:0], pg.plugins...)
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

// IsFunc returns true, if the given [Symbol] is either a function or a
// [NamedSymbol] that in turn also represents a function. For everything else,
// IsFunc returns false.
func IsFunc(s Symbol) bool {
	if namedsym, ok := s.(NamedSymbol); ok {
		return reflect.TypeOf(namedsym.Symbol).Kind() == reflect.Func
	}
	return reflect.TypeOf(s).Kind() == reflect.Func
}
