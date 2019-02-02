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

package plugger

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func plugnames(plugger PluginGroup) []string {
	names := make([]string, len(plugger.plugins))
	for idx, plug := range plugger.plugins {
		names[idx] = plug.Name
	}
	return names
}

var _ = Describe("sorts plugin", func() {

	It("lexicographically", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"alpha", "beta", "gamma"}))
	})

	It("to front", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta"},
				&PluginSpec{Name: "gamma", Placement: "<"},
				&PluginSpec{Name: "alpha"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"gamma", "alpha", "beta"}))
	})

	It("already at front", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "alpha", Placement: "<"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "beta"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"alpha", "beta", "gamma"}))
	})

	It("to back", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta", Placement: ">"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"alpha", "gamma", "beta"}))
	})

	It("already at back", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta"},
				&PluginSpec{Name: "alpha"},
				&PluginSpec{Name: "gamma", Placement: ">"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"alpha", "beta", "gamma"}))
	})

	It("before another named plugin", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha", Placement: "<gamma"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"beta", "alpha", "gamma"}))
	})

	It("before itself", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta", Placement: "<beta"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"alpha", "beta", "gamma"}))
	})

	It("before another named plugin at begin", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha", Placement: "<beta"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"alpha", "beta", "gamma"}))
	})

	It("after another named plugin", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha", Placement: ">beta"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"beta", "alpha", "gamma"}))
	})

	It("after another named plugin at end", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha", Placement: ">gamma"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"beta", "gamma", "alpha"}))
	})

	It("ignores unknown placement", func() {
		plugger := PluginGroup{
			plugins: []*PluginSpec{
				&PluginSpec{Name: "beta", Placement: ">coma"},
				&PluginSpec{Name: "gamma"},
				&PluginSpec{Name: "alpha"},
			},
		}
		plugger.sort()
		Expect(plugnames(plugger)).To(Equal([]string{"alpha", "beta", "gamma"}))
	})

})
