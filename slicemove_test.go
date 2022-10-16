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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("slices", func() {

	DescribeTable("move slice elements",
		func(from, to int, expected []string) {
			s := []string{"A", "B", "C", "D", "E"}
			move(s, from, to)
			Expect(s).To(Equal(expected))
		},
		Entry("don't move", 2, 2, []string{"A", "B", "C", "D", "E"}),
		Entry("forward", 1, 3, []string{"A", "C", "B", "D", "E"}),
		Entry("to end", 1, 5, []string{"A", "C", "D", "E", "B"}),
		Entry("at end", 4, 5, []string{"A", "B", "C", "D", "E"}),
		Entry("rewind", 3, 1, []string{"A", "D", "B", "C", "E"}),
		Entry("at start", 0, 0, []string{"A", "B", "C", "D", "E"}),
	)

})
