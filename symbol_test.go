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
	"io"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type foostruct struct{}

func (f foostruct) String() string { return "foo" }

type fooint int

func (f fooint) String() string { return "foo" }

var _ = Describe("exposed plugin symbols", func() {

	It("validates a correct function Symbol", func() {
		Expect(Symbol[func()]{S: func() {}}.Validate).NotTo(Panic())
	})

	It("validates a correct interface Symbol", func() {
		var b strings.Builder
		Expect(Symbol[io.Writer]{S: &b}.Validate).NotTo(Panic())
		var f foostruct
		Expect(Symbol[fmt.Stringer]{S: f}.Validate).NotTo(Panic())
		Expect(Symbol[fmt.Stringer]{S: &f}.Validate).NotTo(Panic())
		var i fooint
		Expect(Symbol[fmt.Stringer]{S: i}.Validate).NotTo(Panic())
		Expect(Symbol[fmt.Stringer]{S: &i}.Validate).NotTo(Panic())
	})

	It("rejects nil functions and interfaces", func() {
		Expect(Symbol[func()]{S: nil}.Validate).To(PanicWith("func symbol must not be nil"))
		Expect(Symbol[fmt.Stringer]{S: fmt.Stringer(nil)}.Validate).To(PanicWith("interface symbol must not be nil"))
	})

	It("rejects incorrect non-func and non-interface Symbols", func() {
		Expect(Symbol[int]{S: 42}.Validate).To(PanicWith(
			MatchRegexp(`^symbol must be func or interface, but got`)))
	})

	It("completes the plugin name", func() {
		s := Symbol[any]{}
		s.complete(0, runtime.Caller)
		Expect(s.Plugin).To(Equal("go-plugger"))
	})

	It("does not override an already set plugin name", func() {
		const name = "foobarz"
		s := Symbol[any]{Plugin: name}
		s.complete(0, runtime.Caller)
		Expect(s.Plugin).To(Equal(name))
	})

	DescribeTable("panics when unable to determine the plugin name",
		func(outcome string, expected string) {
			s := Symbol[any]{}
			Expect(func() {
				s.complete(0, func(i int) (uintptr, string, int, bool) {
					if outcome == "" {
						return 0, "", 0, false
					}
					return 0, outcome, 0, true
				})
			}).To(PanicWith(expected))
		},
		Entry("caller's file cannot be determined", "", "unable to discover caller for discovering plugin name"),
		Entry("no directory", "foo.bar", "cannot determine plugin name for symbol of type <nil>"),
		Entry("no directory", "/foo.bar", "cannot determine plugin name for symbol of type <nil>"),
	)

})
