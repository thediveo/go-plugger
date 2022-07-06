//go:build plugger_dynamic && dynamicplugintesting

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

package dynamicplugintesting_test

import (
	"testing"

	"github.com/thediveo/go-plugger/v2"
	"github.com/thediveo/go-plugger/v2/dyn"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlugins(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "plugger/test/dynamicplugintesting suite")
}

var _ = Describe("dynamic plugins", func() {

	It("discovery and calls dynamic plugins", func() {
		dyn.Discover(".", true)
		group := plugger.New("dynamicplugintesting")
		Expect(group).NotTo(BeNil())
		pfs := group.Func("PlugFunc")
		Expect(pfs).To(HaveLen(1))
		Expect(pfs[0].(func() string)()).To(Equal("dynfooplug"))
	})

})
