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

// move element in a slice from index from to index to. The to index is allowed
// to point right after the last element, in this case the element from will be
// moved to the end of the slice.
func move[S ~[]E, E any](s S, from, to int) S {
	// I severely miss Python's and Javascript's simplistic way to move elements
	// within slices. Go is just ugly and terrible. Any of its claims to have
	// been inspired by Python is like Steve Balmer claiming to be inspired by
	// Unix (or being a Baseball team manager for what its worth)...
	if from == to {
		return s
	}
	symbol := s[from]
	if from < to {
		// before: [.] [.] [X] [:] [:] [P] [.]
		// after:  [.] [.] [:] [:] [P] [X] [.]

		// border case: after end
		// before: [.] [.] [X] [:] [:] P
		// after:  [.] [.] [:] [:] [X]

		// border case: at end
		// before: [.] [.] [X] P
		// after:  [.] [.] [P]
		to--
		for i := from; i < to; i++ {
			s[i] = s[i+1]
		}
		s[to] = symbol
		return s
	}
	// before: [.] [.] [P] [:] [:] [X] [.]
	// after:  [.] [.] [X] [P] [:] [:] [.]

	// before: [P] [:] [:] [X] [.] [.] [.]
	// after:  [X] [P] [:] [:] [.] [.] [.]
	for i := from; i > to; i-- {
		s[i] = s[i-1]
	}
	s[to] = symbol
	return s
}
