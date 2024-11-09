// Copyright 2024 Harald Albrecht.
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

//go:build go1.23

package plugger

import "iter"

// All returns an iterator over all registered functions/interfaces together
// with their plugin information (name, position hint) in this plugin group.
//
// As plugin groups have no notion of an “index” the All iterator is a two-value
// iterator:
//   - registered function/interface “symbol”
//   - plugin information, such as name and position hint.
//
// Note that this iterator takes a snapshot of the currently registered plugin
// functions and plugin data only at the time when it is called (and not when it
// is created).
func (g *PluginGroup[T]) All() iter.Seq2[T, Symbol[T]] {
	return func(yield func(T, Symbol[T]) bool) {
		for _, plugsym := range g.PluginsSymbols() {
			if !yield(plugsym.S, plugsym) {
				return
			}
		}
	}
}
