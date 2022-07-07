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

var _ = Describe("plugin registration", func() {

	var oldPluginGroups map[string]*PluginGroup

	BeforeEach(func() {
		// First save the old state of the registered plugin groups, and then
		// reset, so each test here runs on an empty plugin group map. This is
		// needed because we will otherwise trash the static plugin example,
		// making it fail depending on the sequence of tests and examples.
		oldPluginGroups = pluginGroups
		pluginGroups = map[string]*PluginGroup{}
	})

	AfterEach(func() {
		pluginGroups = oldPluginGroups
	})

	It("handles correct caller data in registration", func() {
		Expect(func() {
			registerPlugin(
				func(int) (uintptr, string, int, bool) {
					return uintptr(0), "plagueins/foo/plug.go", 0, true
				})
		}).NotTo(Panic())
		plugs := New("plagueins")
		Expect(plugs.Plugins()).To(ConsistOf(HaveField("Name", "foo")))
	})

	When("something is wrong", func() {

		DescribeTable("panics when unable to fetch runtime caller data or with arcane caller data",
			func(file string, ok bool) {
				Expect(func() {
					registerPlugin(
						func(int) (uintptr, string, int, bool) {
							return uintptr(0), file, 0, ok
						})
				}).To(Panic())
			},
			Entry("without caller data", "", false),
			Entry("invalid file name, no package, etc.", "plug.go", true),
			Entry("invalid file name, rooted path without package", "/plug.go", true),
			Entry("invalid file name, unrooted dir without package", "foo/plug.go", true),
		)

		It("panics when attempting to register the same plugin twice", func() {
			Register(WithName("alpha"), WithSymbol(func() {}))
			Expect(func() {
				Register(WithName("alpha"), WithSymbol(func() {}))
			}).To(Panic())
		})

		It("panics when attempting to register the same symbol twice", func() {
			fn := func() {}
			Expect(func() {
				Register(WithName("alpha"), WithSymbol(fn), WithSymbol(fn))
			}).To(Panic())
		})

		It("panics when attempting to register an unnamed named symbol", func() {
			Expect(func() {
				Register(WithName("alpha"), WithNamedSymbol("", func() {}))
			}).To(Panic())
		})

		It("panics when attempting to register something that isn't a function or interface", func() {
			Expect(func() {
				Register(WithName("omega"), WithSymbol(42))
			}).To(Panic())
			Expect(func() {
				omega := 42
				Register(WithName("omega"), WithSymbol(&omega))
			}).To(Panic())
		})

	})

	It("registers named function symbols", func() {
		Register(WithName("alpha"), WithGroup("group"), WithNamedSymbol("Foo", PrefixFoo))
		Register(WithName("beta"), WithGroup("group"), WithNamedSymbol("Foo", PrefixBar))
		plugs := New("group")
		Expect(plugs.Plugins()).To(HaveLen(2))
		Expect(plugs.Func("Foo")).To(HaveLen(2))
	})

	It("registers struct symbols", func() {
		Register(WithName("alpha"), WithGroup("group"), WithSymbol(Ioo(&Loo{})))
		plugs := New("group")
		// Note: the Ioo type is lost on the symbol.
		pis := plugs.PluginsFunc("Loo")
		Expect(pis).To(ConsistOf(HaveField("F", And(
			BeAssignableToTypeOf(Ioo(&Loo{})),
			WithTransform(func(f Ioo) int { return f.Goo() }, Equal(42)),
		))))
	})

	It("registers named struct symbols", func() {
		Register(WithName("alpha"), WithGroup("group"), WithNamedSymbol("Ioo", Ioo(&Loo{})))
		plugs := New("group")
		pis := plugs.PluginsFunc("Ioo")
		Expect(pis).To(ConsistOf(HaveField("F", And(
			BeAssignableToTypeOf(Ioo(&Loo{})),
			WithTransform(func(f Ioo) int { return f.Goo() }, Equal(42)),
		))))
	})

})

func Foo()       {}
func PrefixFoo() {}
func PrefixBar() {}

type Ioo interface {
	Goo() int
}

type Loo struct{}

var _ Ioo = (*Loo)(nil)

func (l *Loo) Goo() int { return 42 }
