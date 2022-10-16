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

/*
Package example demonstrates how to register plugins and how to work with the
exposed plugin symbols, such as calling exposed plugin functions.
*/
package example

import (
	"fmt"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/go-plugger/v3/example/plugin"

	_ "github.com/thediveo/go-plugger/v3/example/barplug"
	_ "github.com/thediveo/go-plugger/v3/example/fooplug"
)

// Retrieves the (ordered) list of exposed symbols of type [plugin.DoItFn] and
// then calls each exposed DoIt functions one after another, printing out the
// string returned by each.
//
// # Note
//
// The plugin interface, that is, the exposed symbol type(s) in this example are
// defined in a separate package in order to avoid import cycles.
func Example() {
	doIts := plugger.Group[plugin.DoItFn]()
	for _, doIt := range doIts.Symbols() {
		fmt.Println(doIt())
	}
	// Output: fooplug static plugin
	// barplug static plugin
}
