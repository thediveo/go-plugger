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

// RegisterOption configures [PluginSpec] aspects for registering plugins using
// [Register].
type RegisterOption func(*PluginSpec)

// WithName specifies the name of the plugin to be registered.
func WithName(name string) RegisterOption {
	return func(ps *PluginSpec) {
		ps.Name = name
	}
}

// WithGroup specifies the group name to register the plugin in.
func WithGroup(group string) RegisterOption {
	return func(ps *PluginSpec) {
		ps.Group = group
	}
}

// WithPlacement specifies the placement of the plugin to be registered in
// relation to other plugins. The placement can take on one of the following
// forms:
//   - "<": before all already registered plugins.
//   - ">": after all already registered plugins.
//   - "<foo": before the already registered plugin named "foo".
//   - "<bar": after the already registered plugin named "bar".
func WithPlacement(placement string) RegisterOption {
	return func(ps *PluginSpec) {
		ps.Placement = placement
	}
}

// WithSymbol adds a plugin exported function. WithSymbol can be used multiple
// times to register multiple plugin function, and can also be freely mixed with
// [WithNamedSymbol].
func WithSymbol(fn Symbol) RegisterOption {
	return func(ps *PluginSpec) {
		ps.Symbols = append(ps.Symbols, fn)
	}
}

// WithNamedSymbol adds a named symbol of a plugin exported function.
// WithNamedSymbol can be used multiple times in order to register multiple
// plugin functions, as well as freely mixed with [WithFunction].
func WithNamedSymbol(name string, symbol Symbol) RegisterOption {
	return func(ps *PluginSpec) {
		ps.Symbols = append(ps.Symbols, NamedSymbol{
			Name:   name,
			Symbol: symbol,
		})
	}
}
