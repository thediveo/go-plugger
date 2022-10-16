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
	"os"
	"path/filepath"
	"reflect"
)

// Symbol is a function or interface exposed by a (named) plugin. The interface
// must not be a constraint interface used to express type constraints.
//
// The placement hint indicates where in an ordered list of the plugin symbols
// this plugin should be placed:
//   - "<": place at the beginning;
//   - ">": place at the end;
//   - "<foo": place before the plugin named "foo", if there is no such plugin
//     named "foo", then the placement gets ignored;
//   - ">foo": place after the plugin named "foo", if there is no such plugin
//     named "foo", then the placement gets ignored.
type Symbol[T any] struct {
	S         T      // exposed function or interface symbol.
	Plugin    string // name of plugin exposing the symbol S.
	Placement string // optional placement hint, or "".
}

type symbolSetter interface {
	setPlugin(name string)
	setPlacement(placement string)
	complete(offset int, runtimeCaller func(int) (uintptr, string, int, bool))
}

var _ symbolSetter = (*Symbol[any])(nil)

// Validate an exported plugin symbol and panic if the symbol is anything other
// than a function or interface.
//
// While Go 1 has gained type constraints (in form of constraint interfaces) for
// use with Generics, there currently is no way to express constraints that
// forbid certain types, instead of allowing only a specific set. Thus, we need
// to validate at runtime that the symbol's type T actually is either a function
// type or an interface type. However, we cannot simply query the type of the
// symbol as this in the case of T being an interface would return the
// implementing value's T*. We thus need to construct a dummy composite type
// containing T that reflect accepts and then get that contained T's type via
// reflect. This then will be the correct interface T (instead of the underlying
// implementing value's T*). The Go compiler already ensured that the value
// satisfies the interface type T.
func (s Symbol[T]) Validate() {
	var dummyCompositeT []T // https://stackoverflow.com/a/18316266
	switch reflect.TypeOf(dummyCompositeT).Elem().Kind() {
	case reflect.Func:
		if reflect.ValueOf(s.S).IsNil() {
			panic("func symbol must not be nil")
		}
	case reflect.Interface:
		v := reflect.ValueOf(s.S)
		if v.Kind() == reflect.Invalid || (v.Kind() == reflect.Pointer && v.IsNil()) {
			panic("interface symbol must not be nil")
		}
	default:
		panic(fmt.Sprintf("symbol must be func or interface, but got %T", s.S))
	}
}

// sets the plugin name of an exposed symbol.
func (s *Symbol[T]) setPlugin(name string) {
	s.Plugin = name
}

// sets the placement hint of an exposed symbol.
func (s *Symbol[T]) setPlacement(placement string) {
	s.Placement = placement
}

// completes the blanks, that is, fills in the plugin name derived from the
// directory name of the package of the original caller (taking offset into
// account).
func (s *Symbol[T]) complete(offset int, runtimeCaller func(int) (uintptr, string, int, bool)) {
	if s.Plugin != "" {
		return
	}
	_, file, _, ok := runtimeCaller(offset + 1)
	if !ok {
		panic("unable to discover caller for discovering plugin name")
	}
	s.Plugin = filepath.Base(filepath.Dir(file))
	switch s.Plugin {
	case "", ".", string(os.PathSeparator):
		panic(fmt.Sprintf("cannot determine plugin name for symbol of type %T", s.S))
	}
}
