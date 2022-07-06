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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testGroupName = "foobar"
const emptyGroupName = "empty-foobar"

// tregister is a minimal non-checking surrogate for Register for use solely in
// tests. For testing purposes, allow registering "things" that aren't
// functions.
func tregister(opts ...RegisterOption) {
	var pspec PluginSpec
	for _, opt := range opts {
		opt(&pspec)
	}
	pspec.symbolmap = map[string]Symbol{}
	for _, symbol := range pspec.Symbols {
		var symname string
		var sym Symbol
		func() {
			defer func() {
				if r := recover(); r != nil {
					sym = symbol
				}
			}()
			symname, sym = resolveSymbol(&pspec, symbol)
		}()
		pspec.symbolmap[symname] = sym
	}
	pg := New(pspec.Group)
	pg.plugins = append(pg.plugins, &pspec)
	pg.unordered = true
}

var _ = Describe("plugin groups", func() {

	var oldPluginGroups map[string]*PluginGroup

	BeforeEach(func() {
		// First save the old state of the registered plugin groups, and then
		// reset, so each test here runs on an empty plugin group map. This is
		// needed because we will otherwise trash the static plugin example,
		// making it fail depending on the sequence of tests and examples.
		oldPluginGroups = pluginGroups
		pluginGroups = map[string]*PluginGroup{}

		tregister(WithName("plug-a"), WithGroup(testGroupName),
			WithNamedSymbol("DoIt", func() string { return "DoIt plug-a" }))
		tregister(WithName("plug-b"), WithGroup(testGroupName),
			WithNamedSymbol("DoIt", 42))
		tregister(WithName("plug-c"), WithGroup(testGroupName),
			WithNamedSymbol("DoIt", func() string { return "DoIt plug-c" }))
		tregister(WithName("plug-d"), WithGroup(testGroupName),
			WithNamedSymbol("DontYouDoIt", func() string { return "DontYouDoIt plug-d" }))

		DeferCleanup(func() {
			pluginGroups = oldPluginGroups
		})
	})

	It("always returns the same PluginGroup for a given name", func() {
		pgOne := New(emptyGroupName)
		Expect(pgOne).NotTo(BeNil())
		pgTwo := New(emptyGroupName)
		Expect(pgTwo).To(BeIdenticalTo(pgOne))
	})

	Context("querying plugin-exported functions", func() {

		It("returns a function symbol", func() {
			fns := New(testGroupName).Func("DoIt")
			Expect(fns).To(HaveLen(2))
			Expect(fns).To(HaveEach(BeAssignableToTypeOf(func() string { return "" })))
		})

		It("returns function symbols matching prefix", func() {
			fns := New(testGroupName).FuncPrefix("Do")
			Expect(fns).To(HaveLen(3))
			Expect(fns).To(HaveEach(BeAssignableToTypeOf(func() string { return "" })))
		})

		It("returns exported functions", func() {
			fns := New(testGroupName).PluginsFunc("DoIt")
			Expect(fns).To(HaveLen(2))
			Expect(fns).To(HaveEach(And(
				HaveField("Name", Not(BeEmpty())),
				HaveField("Plugin", HaveField("Name", Not(BeEmpty()))),
				HaveField("F", BeAssignableToTypeOf(func() string { return "" })),
			)))
		})

		It("returns a specific plugins exported function", func() {
			fn := New(testGroupName).PluginFunc("plug-c", "DoIt")
			Expect(fn).NotTo(BeNil())

		})

		It("returns registered plugins", func() {
			plugs := New(testGroupName).Plugins()
			Expect(plugs).To(HaveLen(4))
			tregister(WithName("plug-e"), WithGroup(testGroupName),
				WithNamedSymbol("DontYouDoIt", func() string { return "DontYouDoIt plug-e" }))
			Expect(plugs).To(HaveLen(4))
			plugs = New(testGroupName).Plugins()
			Expect(plugs).To(HaveLen(5))
		})

		It("returns the names of registered plugins in a group", func() {
			Expect(New(testGroupName).PluginNames()).To(ConsistOf(
				"plug-a", "plug-b", "plug-c", "plug-d",
			))
		})

	})

	Context("error resilience", func() {

		It("handles non-existing symbols", func() {
			group := New(testGroupName)
			Expect(group.Func("XXX")).To(BeEmpty())
			Expect(group.FuncPrefix("XXX")).To(BeEmpty())
			Expect(group.PluginsFunc("XXX")).To(BeEmpty())
			Expect(group.PluginFunc("XXX", "XXX")).To(BeNil())
			Expect(group.PluginFunc("plug-a", "XXX")).To(BeNil())
		})

	})

	Context("plugin order", func() {
		const orderGroupName = "test-ooorder"

		It("orders lexicographically", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

		It("places at the beginning", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName))
			tregister(WithName("gamma"), WithGroup(orderGroupName), WithPlacement("<"))
			tregister(WithName("alpha"), WithGroup(orderGroupName))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"gamma", "alpha", "beta"}))
		})

		It("places the first at the beginning", func() {
			group := New(orderGroupName)
			tregister(WithName("alpha"), WithGroup(orderGroupName), WithPlacement("<"))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("beta"), WithGroup(orderGroupName))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

		It("places at the end", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName), WithPlacement(">"))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "gamma", "beta"}))
		})

		It("places the last at the end", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName))
			tregister(WithName("gamma"), WithGroup(orderGroupName), WithPlacement(">"))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

		It("places before another named plugin", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName), WithPlacement("<gamma"))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"beta", "alpha", "gamma"}))
		})

		It("places before another named plugin at the beginning", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName), WithPlacement("<beta"))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

		It("places itself before itself", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName), WithPlacement("<beta"))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

		It("places after another named plugin", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName), WithPlacement(">beta"))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"beta", "alpha", "gamma"}))
		})

		It("places after another named plugin at the end", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName), WithPlacement(">gamma"))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"beta", "gamma", "alpha"}))
		})

		It("places itself after itself", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName), WithPlacement(">beta"))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

		It("ignores an unknown placement", func() {
			group := New(orderGroupName)
			tregister(WithName("beta"), WithGroup(orderGroupName), WithPlacement(">coma"))
			tregister(WithName("gamma"), WithGroup(orderGroupName))
			tregister(WithName("alpha"), WithGroup(orderGroupName))
			group.sort()
			Expect(group.PluginNames()).To(Equal([]string{"alpha", "beta", "gamma"}))
		})

	})

	DescribeTable("knows what a function symbol is",
		func(s Symbol, expected bool) {
			Expect(IsFunc(s)).To(Equal(expected))
		},
		Entry("not a symbol", 42, false),
		Entry("a func", func() {}, true),
		Entry("a named symbol", NamedSymbol{"foo", func() {}}, true),
	)

})
