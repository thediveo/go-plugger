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

// Package foo is a static plugin skeleton demonstrating the plugger
// registration mechanism. This plugin automatically registers itself as a
// plugin if it gets imported as a package into an application. For
// demonstration purposes, a single DoIt() plugin function gets exported as
// part of the self-registration.
//
// Example Plugin Self-Registration
//
// A typical way to carry out the plugin self-registration is with an init()
// function, such as the following one. This example exports a single plugin
// function named "DoIt". The plugin name and its plugin group will be
// autodiscovered, because we're leaving the particular fields in a PluginSpec
// left unspecified and thus zeroed. The plugin name thus will be the name
// (but not path) of the directory this plugin package is located in. And the
// plugin group will be the name of the directory where the plugin directory
// or directories are in.
//
//   func init() {
//       plugger.RegisterPlugin(&plugger.PluginSpec{
//           Symbols: []plugger.Symbol{DoIt},
//       })
//   }
//
// As a sidenote, this same self-registration mechanism can be kept as is when
// later deciding to convert a plugin from being statically linked into its
// application to a shared library-based deployment using .so plugins.
package foo

import plugger "github.com/thediveo/go-plugger"

// Register this plugin with its exported plugin function(s). The plugin name
// and group are left zeroed, so they will be discovered automatically: the
// plugin name is taken from this (plugin) package's directory name (not:
// path). And the plugin group name is taken from the directory name where the
// plugin directory (or directories) is/are in. These autodetected data can be
// overriden by explicitly specifying them here in the PluginSpec given to the
// call of RegisterPlugin.
func init() {
	plugger.RegisterPlugin(&plugger.PluginSpec{
		Symbols: []plugger.Symbol{DoIt},
	})
}

// DoIt is an exemplary exported plugin function. As usual, it needs to be
// public and it has to be exported using a PluginSpec given to the call of
// RegisterPlugin.
func DoIt() (result string) {
	return "foo static plugin"
}
