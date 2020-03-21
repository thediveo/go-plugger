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

package plugger_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	plugger "github.com/thediveo/go-plugger"
	barplug "github.com/thediveo/go-plugger/internal/staticplugintesting/barplug"
	fooplug "github.com/thediveo/go-plugger/internal/staticplugintesting/fooplug"
	zooplug "github.com/thediveo/go-plugger/internal/staticplugintesting/zooplug"
)

var _ = Describe("static plugins", func() {

	It("register with deduced/set names and group", func() {
		// registering order differs from alphabetical order, and placement.
		zooplug.DoRegister()
		barplug.DoRegister()
		fooplug.DoRegister()
		plugs := plugger.New("staticplugintesting").Plugins()
		Expect(plugs).To(HaveLen(3))
		Expect(*plugs[0]).To(MatchFields(IgnoreExtras, Fields{
			"Name":      Equal("fooplug"),
			"Group":     Equal("staticplugintesting"),
			"Placement": Equal("<barplug"),
		}))
		Expect(*plugs[1]).To(MatchFields(IgnoreExtras, Fields{
			"Name":      Equal("barplug"),
			"Group":     Equal("staticplugintesting"),
			"Placement": Equal(""),
		}))
		Expect(*plugs[2]).To(MatchFields(IgnoreExtras, Fields{
			"Name":      Equal("zoo"),
			"Group":     Equal("staticplugintesting"),
			"Placement": Equal(""),
		}))
	})

	It("finds the exported named function", func() {
		plugfuncs := plugger.New("staticplugintesting").Func("PlugFunc")
		Expect(plugfuncs).To(HaveLen(3))
		r := make([]string, 0, 3)
		for _, f := range plugfuncs {
			r = append(r, f.(func() string)())
		}
		Expect(r).To(Equal([]string{
			"fooplug", "barplug", "zooplug",
		}))
	})

	It("finds the exported named function with plugin information", func() {
		plugfuncs := plugger.New("staticplugintesting").PluginsFunc("PlugFunc")
		Expect(plugfuncs).To(HaveLen(3))
		Expect(plugfuncs).To(MatchElements(
			func(e interface{}) string { return "*" },
			AllowDuplicates,
			Elements{
				"*": MatchFields(IgnoreExtras, Fields{
					"Name": Equal("PlugFunc"),
				}),
			},
		))
		Expect(plugfuncs[0].Plugin.Name).To(Equal("fooplug"))
		Expect(plugfuncs[1].Plugin.Name).To(Equal("barplug"))
		Expect(plugfuncs[2].Plugin.Name).To(Equal("zoo"))
	})

	It("locates an exported named function in a specific plugin", func() {
		p := plugger.New("staticplugintesting")

		Expect(p.PluginFunc("noneplug", "PlugFunc")).To(BeNil())
		Expect(p.PluginFunc("zoo", "NonSuchFunction")).To(BeNil())

		zooplugfunc := p.PluginFunc("zoo", "PlugFunc")
		Expect(zooplugfunc).NotTo(BeNil())
		Expect(reflect.ValueOf(zooplugfunc).Pointer()).To(
			Equal(reflect.ValueOf(zooplug.PlugFunc).Pointer()))
	})

	It("returns fake plugger for unknown group", func() {
		p := plugger.New("unknown")
		Expect(p.Group).To(Equal("unknown"))
		Expect(p.Plugins()).To(HaveLen(0))
	})

})
