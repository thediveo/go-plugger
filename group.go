// Copyright 2022 Harald Albrecht.
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
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/exp/slices"
)

// PluginGroup represents the exposed plugin symbols for a particular symbol
// type, with the exposed symbols ordered by plugin name, or alternatively, by
// plugin placement.
type PluginGroup[T any] struct {
	mu      sync.RWMutex // protects the following elements.
	ordered bool         // has the list of registered plugin symbols been ordered or is it still unordered?
	symbols []Symbol[T]  // (ordered) list of registered plugin symbols.
}

// GroupStash is a “backup” of a PluginGroup. It can be used especially in
// unit tests where a PluginGroup needs to be modified to a particular known
// configuration for a test, and the group's original configuration restored
// after the test.
type GroupStash[T any] struct {
	ordered bool
	symbols []Symbol[T]
}

// Group returns the [*PluginGroup] object for the given exposed symbol type T.
// Calling Group multiple times for the same exposed symbol type T always
// returns the same [PluginGroup] object.
func Group[T any]() *PluginGroup[T] {
	var dummyCompositeT []T // https://stackoverflow.com/a/18316266
	t := reflect.TypeOf(dummyCompositeT).Elem()
	groupsmu.Lock()
	defer groupsmu.Unlock()
	group := groups[t]
	if group == nil {
		group = &PluginGroup[T]{}
		groups[t] = group
	}
	return group.(*PluginGroup[T])
}

// groups maps function and interface types to their (typed) plugin groups.
var groupsmu sync.Mutex
var groups = map[reflect.Type]any{} // actually, *PluginGroup[T]

// String renders a textual representation of a particular Group, showing the
// managed symbol type as well as the plugin-exposed symbols registered in this
// group.
func (g *PluginGroup[T]) String() string {
	g.lock()
	defer g.unlock()

	var s strings.Builder
	s.WriteString("PluginGroup[")
	var dummyCompositeT []T // https://stackoverflow.com/a/18316266
	symbolType := reflect.TypeOf(dummyCompositeT).Elem()
	s.WriteString(symbolType.PkgPath())
	s.WriteRune('.')
	s.WriteString(symbolType.Name())
	s.WriteString("]: [")
	for idx, symbol := range g.symbols {
		if idx > 0 {
			s.WriteRune(',')
		}
		s.WriteRune('"')
		s.WriteString(symbol.Plugin)
		s.WriteString(`":`)
		if fn := runtime.FuncForPC(reflect.ValueOf(symbol.S).Pointer()); fn != nil {
			s.WriteString(fn.Name())
		} else {
			s.WriteString(fmt.Sprintf("%#v", symbol.S))
		}
	}
	s.WriteRune(']')
	return s.String()
}

// RegisterOption allows optional registration information to be passed to the
// Register method of plugin groups.
type RegisterOption func(symbolSetter)

// Register a plugin-exposed symbol, with optional additional registration
// information.
func (g *PluginGroup[T]) Register(symbol T, opts ...RegisterOption) {
	s := Symbol[T]{S: symbol}
	s.Validate() // panics if mistreated to a non-function and non-interface type symbol.
	s.complete(1, runtime.Caller)
	for _, option := range opts {
		option(&s)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ordered = false
	g.symbols = append(g.symbols, s)
}

// WithPlugin registers an exposed symbol with the given plugin name in
// [plugger.PluginGroup.Register].
func WithPlugin(name string) func(symbolSetter) {
	return func(s symbolSetter) {
		s.setPlugin(name)
	}
}

// WithPlacement registers an exposed symbol with the given (plugin) placement
// hint in [plugger.PluginGroup.Register].
func WithPlacement(placement string) func(symbolSetter) {
	return func(s symbolSetter) {
		s.setPlacement(placement)
	}
}

// Symbols returns all symbols (functions or interfaces) exposed by the plugins
// in this Group. This is always a clean and ordered copy of the list of exposed
// symbols.
func (g *PluginGroup[T]) Symbols() []T {
	g.lock()
	defer g.unlock()

	s := make([]T, 0, len(g.symbols))
	for _, symbol := range g.symbols {
		s = append(s, symbol.S)
	}
	return s
}

// PluginsSymbols returns all exposed symbols together with the names of the
// plugins exposing them. This is always a clean and ordered copy of the
// [Symbol] objects.
func (g *PluginGroup[T]) PluginsSymbols() []Symbol[T] {
	g.lock()
	defer g.unlock()

	return slices.Clone(g.symbols)
}

// PluginSymbol returns the exposed symbol of the plugin identified by its name,
// or the zero symbol value if no such named plugin exists in this symbol group.
func (g *PluginGroup[T]) PluginSymbol(name string) T {
	g.lock()
	defer g.unlock()

	for _, symbol := range g.symbols {
		if symbol.Plugin == name {
			return symbol.S
		}
	}
	var zero T
	return zero
}

// Plugins returns the names of all plugins exposing symbols in this plugin
// group. The returned list is always ordered, based on the plugin names and
// placement hints.
func (g *PluginGroup[T]) Plugins() []string {
	g.lock()
	defer g.unlock()

	plugins := make([]string, 0, len(g.symbols))
	for _, symbol := range g.symbols {
		plugins = append(plugins, symbol.Plugin)
	}
	return plugins
}

// Clears this plugin group's configuration (such as in unit tests).
func (g *PluginGroup[T]) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ordered = false
	g.symbols = nil
}

// Save returns a copy of this plugin group's current plugin configuration, for
// later restoration using the Restore method.
func (g *PluginGroup[T]) Backup() GroupStash[T] {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return GroupStash[T]{
		ordered: g.ordered,
		symbols: slices.Clone(g.symbols),
	}
}

// Restore a plugin group's former plugin configuration from a backup previously
// created by the Backup method.
func (g *PluginGroup[T]) Restore(s GroupStash[T]) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ordered = s.ordered
	g.symbols = slices.Clone(s.symbols)
}

// sort the plugins by name and optionally by reference; that is, individual
// plugins can claim to get to the front/end, or before/after a another named
// plugin. This method must be called under write lock.
//
// The plugin ordering mechanism is with a nod to Jeremy Ruston and his
// incredible TiddlyWiki (in particular, its list and module sorting).
func (g *PluginGroup[T]) sort() {
	// First, sort lexicographically by plugin name (not: by plugin path).
	sort.Slice(g.symbols, func(a, b int) bool {
		return g.symbols[a].Plugin < g.symbols[b].Plugin
	})
	// Second, honor the optional positional requests of individual plugins.
	// Or, at least try to do so...
	symbols := slices.Clone(g.symbols)
	for _, symbol := range g.symbols {
		// Find the next plugin to process from the original list on in the
		// current and potentially modified list, because we need to work on the
		// current list when shuffling plugins around.
		var idx int
		var sym Symbol[T]
		for idx, sym = range symbols {
			if sym.Plugin == symbol.Plugin {
				break
			}
		}
		pos := idx // start with no change in a plugin's sequence position
		// Does the plugin want to be positioned either before a specifically
		// named other plugin or at the beginning?
		if strings.HasPrefix(symbol.Placement, "<") {
			before := symbol.Placement[1:]
			if before == "" {
				pos = 0 // tangarines FIRST (*all* of them, *snicker*)
			} else {
				// Find the named plugin at its current position; not at the
				// original position, that wouldn't make sense and mix up the
				// original intention.
				for i, p := range symbols {
					if before == p.Plugin {
						pos = i
						break
					}
				}
			}
		}
		// Does the plugin want to be positioned either after another
		// specifically named plugin or at the end of the sequence?
		if strings.HasPrefix(symbol.Placement, ">") {
			after := symbol.Placement[1:]
			if after == "" {
				pos = len(symbols)
			} else {
				// Find the named plugin at its current position; not at the
				// original position, that wouldn't make sense and mix up the
				// original intention.
				for i, p := range symbols {
					if after == p.Plugin {
						pos = i + 1
						break
					}
				}
			}
		}
		symbols = move(symbols, idx, pos)
	}
	g.symbols = symbols
}

// lock locks the plugin group against concurrent write changes and sorts the
// plugin exposed list of symbols, if necessary. The caller needs to (defer to)
// unlock after having done its work.
func (g *PluginGroup[T]) lock() {
	g.mu.RLock()
	// As we cannot downgrade a write lock into a read lock atomatically, we
	// need to rinse and repeat until got our read lock on a sorted exposed
	// plugin symbols list...
	for !g.ordered { // https://github.com/golang/go/issues/4026#issuecomment-66069822
		g.mu.RUnlock()
		// Here, another goroutine might win an unintended race with us to sort
		// the list of exposed plugin symbols, so skip the sort operation if we
		// finally got the write lock on a sorted list.
		g.mu.Lock()
		if !g.ordered {
			g.sort()
			g.ordered = true
		}
		g.mu.Unlock()
		// Here, the list might get unsorted again if we're unlucky.
		g.mu.RLock()
	}
}

// unlock unlocks the plugin group.
func (g *PluginGroup[T]) unlock() {
	g.mu.RUnlock()
}
