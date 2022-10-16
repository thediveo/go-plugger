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
Package fooplug is an example plugin registering its exposed DoIt function
symbol.
*/
package fooplug

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/go-plugger/v3/example/plugin"
)

// DoIt is an exposed plugin symbol.
func DoIt() string { return "fooplug static plugin" }

// Typesafe registration of our exposed plugin symbol with a twist: we want our
// plugin (symbol) to be ordered before barplug.
func init() {
	plugger.Group[plugin.DoItFn]().Register(DoIt, plugger.WithPlacement("<barplug"))
}
