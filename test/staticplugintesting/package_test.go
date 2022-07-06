// Copyright 2021, 2022 Harald Albrecht.
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

package staticplugintesting_test

import (
	"testing"

	"github.com/thediveo/go-plugger/v2"
	"github.com/thediveo/go-plugger/v2/test/staticplugintesting/barplug"
	"github.com/thediveo/go-plugger/v2/test/staticplugintesting/fooplug"
	"github.com/thediveo/go-plugger/v2/test/staticplugintesting/zooplug"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestPlugins(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "plugger/test/staticplugintesting suite")
}

var _ = Describe("static plugins", func() {

	It("register themselves", func() {
		zooplug.DoRegister()
		fooplug.DoRegister()
		barplug.DoRegister()

		group := plugger.New("staticplugintesting")
		Expect(group).NotTo(BeNil())
		Expect(group.Plugins()).To(ContainElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":    Equal("barplug"),
				"Group":   Equal("staticplugintesting"),
				"Symbols": HaveLen(1),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":    Equal("fooplug"),
				"Group":   Equal("staticplugintesting"),
				"Symbols": HaveLen(1),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":    Equal("zoo"), // sic!
				"Group":   Equal("staticplugintesting"),
				"Symbols": HaveLen(1),
			})),
		))

		out := []string{}
		Expect(group.Func("PlugFunc")).To(HaveLen(3))
		for _, fn := range group.Func("PlugFunc") {
			out = append(out, fn.(func() string)())
		}
		Expect(out).To(ContainElements("barplug", "fooplug", "zooplug"))
	})

})
