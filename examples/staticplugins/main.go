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

// Demonstrates how to statically include self-registering plugins (contained
// inside the plugins/ subdirectory) and how to call their exported functions.
//
// Example Code
//
// First, import the plugins (packages) that are going to be part of your
// application. The import needs to be only done using the blank identifier,
// as we just need to pull in the plugin(s). This will trigger the plugins to
// also self-register, so we can use them later in our app.
//
//   import _ "github.com/thediveo/go-plugger/examples/staticplugins/plugins/foo"
//
// The plugger plugin management groups plugins – in order to allow multiple
// sets of plugins for different functionalities between groups. In our
// example, the plugins will self-register into the (witlessly named)
// "plugins" group. We thus ask for this plugin group "plugins", then iterate
// over each of the plugins belonging to it, calling each plugin's DoIt()
// functionality.
//
//   plugs := plugger.New("plugins")
//   for _, doit := range plugs.Func("DoIt") {
//       fmt.Println(doit.(func() string)())
//   }
package main

import (
	"fmt"

	plugger "github.com/thediveo/go-plugger"
	_ "github.com/thediveo/go-plugger/examples/staticplugins/plugins/foo"
)

func main() {
	plugs := plugger.New("plugins")
	for _, doit := range plugs.Func("DoIt") {
		fmt.Println(doit.(func() string)())
	}
}
