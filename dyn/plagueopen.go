//go:build plugger_dynamic

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

package dyn

import "plugin"

// This, erm, "plugs in" the plugin.Open implementation only when the build
// tag/constraint plugger_dynamic has been specified. This prevents the Go
// linker getting berserk when building static Go binaries without the dynamic
// plugin loading required; otherwise the Go linker will complain as soon as the
// plug.Open symbol is being present (even if not used at all) and a static
// binary is to be build.
func init() {
	pluginOpen = func(path string) error {
		_, err := plugin.Open(path)
		return err
	}
}
