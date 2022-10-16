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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// two separate (plugin) function types results in two separate (plugin) groups.
type fooFn func() string
type barFn func() string
type fooIf interface {
	Foo() string
}

type fooImpl struct{ s string }

func (f fooImpl) Foo() string { return f.s }

var _ = Describe("exposed plugin symbol groups", func() {

	BeforeEach(func() {
		groups = map[reflect.Type]any{}
	})

	Context("concurrency-safe", func() {

		It("always returns the same plugin group for a specific type", func() {
			ch := make(chan *PluginGroup[fooFn])
			for i := 0; i < 2; i++ {
				go func() {
					ch <- Group[fooFn]()
				}()
			}
			// Nota bene: BeIndenticalTo won't ever compare nil to nil
			Expect(<-ch).To(BeIdenticalTo(<-ch))
		})

	})

	It("renders a textual representation of the type and exposed symbols", func() {
		fooIfGroup := Group[fooIf]()
		fooIfGroup.Register(&fooImpl{s: "one"}, WithPlugin("one"))
		fooIfGroup.Register(&fooImpl{s: "two"}, WithPlugin("two"), WithPlacement("<"))
		for i := 0; i < 2; i++ {
			Expect(fmt.Sprintf("%s", fooIfGroup)).To(MatchRegexp(
				`PluginGroup\[github\.com/thediveo/go-plugger/v3\.fooIf\]: \["two":.*,"one":.*\]`))
		}

		barFnGroup := Group[barFn]()
		barFnGroup.Register(func() string { return "one" }, WithPlugin("one"))
		barFnGroup.Register(func() string { return "two" }, WithPlugin("two"), WithPlacement("<one"))
		Expect(fmt.Sprintf("%s", barFnGroup)).To(MatchRegexp(
			`PluginGroup\[github\.com/thediveo/go-plugger/v3\.barFn\]: \["two":.*\.\.func.*,"one":.*\.\.func.*\]`))
	})

	It("doesn't mix exported symbol types", func() {
		fooGroup := Group[fooFn]()
		Expect(fooGroup).NotTo(BeNil())
		barGroup := Group[barFn]()
		Expect(barGroup).NotTo(BeNil())
		Expect(fooGroup).NotTo(BeIdenticalTo(barGroup))
	})

	It("registers symbols and sorts them", func() {
		g := Group[fooFn]()
		Expect(g).NotTo(BeNil())
		g.Register(func() string { return "one" }, WithPlugin("one"))
		g.Register(func() string { return "two" }, WithPlugin("two"), WithPlacement("<one"))
		Expect(g.Plugins()).To(ConsistOf("two", "one"))
		syms := g.Symbols()
		Expect(syms).To(HaveLen(2))
		Expect([]string{
			syms[0](), syms[1](),
		}).To(ContainElements("one", "two"))
		Expect(g.PluginsSymbols()).To(ConsistOf(
			HaveField("Plugin", "two"),
			HaveField("Plugin", "one"),
		))
	})

	It("finds a specific plugin's symbol", func() {
		g := Group[fooFn]()
		Expect(g).NotTo(BeNil())
		g.Register(func() string { return "one" }, WithPlugin("one"))
		Expect(g.PluginSymbol("foo")).To(BeNil())
		foofn := g.PluginSymbol("one")
		Expect(foofn).NotTo(BeNil())
		Expect(foofn()).To(Equal("one"))
	})

	It("fills in the plugin name if missing", func() {
		g := Group[fooFn]()
		Expect(g).NotTo(BeNil())
		g.Register(func() string { return "one" })
		Expect(g.PluginsSymbols()).To(HaveEach(HaveField("Plugin", "go-plugger")))
	})

	DescribeTable("orders plugins",
		func(a, ap, b, bp, c, cp string, expected []string) {
			g := &PluginGroup[any]{
				symbols: []Symbol[any]{
					{Plugin: a, Placement: ap},
					{Plugin: b, Placement: bp},
					{Plugin: c, Placement: cp},
				},
			}
			g.sort()
			Expect(g.Plugins()).To(Equal(expected))
		},
		Entry("lexicographically",
			"beta", "", "gamma", "", "alpha", "",
			[]string{"alpha", "beta", "gamma"}),
		Entry("places at the beginning",
			"beta", "", "gamma", "<", "alpha", "",
			[]string{"gamma", "alpha", "beta"}),
		Entry("places the first at the beginning",
			"alpha", "<", "gamma", "", "beta", "",
			[]string{"alpha", "beta", "gamma"}),
		Entry("places at the end",
			"beta", ">", "gamma", "", "alpha", "",
			[]string{"alpha", "gamma", "beta"}),
		Entry("places the last at the end",
			"beta", "", "alpha", "", "gamma", ">",
			[]string{"alpha", "beta", "gamma"}),
		Entry("places before another named plugin",
			"beta", "", "gamma", "", "alpha", "<gamma",
			[]string{"beta", "alpha", "gamma"}),
		Entry("places before another named plugin at the beginning",
			"beta", "", "gamma", "", "alpha", "<beta",
			[]string{"alpha", "beta", "gamma"}),
		Entry("places itself before itself",
			"beta", "<beta", "gamma", "", "alpha", "",
			[]string{"alpha", "beta", "gamma"}),
		Entry("places after another named plugin",
			"beta", "", "gamma", "", "alpha", ">beta",
			[]string{"beta", "alpha", "gamma"}),
		Entry("places after another named plugin at the end",
			"beta", "", "gamma", "", "alpha", ">gamma",
			[]string{"beta", "gamma", "alpha"}),
		Entry("places itself after itself",
			"beta", ">beta", "gamma", "", "alpha", "",
			[]string{"alpha", "beta", "gamma"}),
		Entry("ignores an unknown placement",
			"beta", ">coma", "gamma", "", "alpha", "",
			[]string{"alpha", "beta", "gamma"}),
	)

})
