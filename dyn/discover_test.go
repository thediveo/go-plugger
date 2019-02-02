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

package dyn

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	plugger "github.com/thediveo/go-plugger"
)

type mockedFileInfo struct {
	name  string
	isdir bool
}

func (mfi mockedFileInfo) Name() string       { return mfi.name }
func (mfi mockedFileInfo) Size() int64        { return 42 }
func (mfi mockedFileInfo) Mode() os.FileMode  { return 0 }
func (mfi mockedFileInfo) ModTime() time.Time { return time.Time{} }
func (mfi mockedFileInfo) IsDir() bool        { return mfi.isdir }
func (mfi mockedFileInfo) Sys() interface{}   { return nil }

var _ = Describe("dynamic plugin", func() {

	Describe("dynamic plugin registering", func() {

		It("discovers nothing in test plugin dir itself", func() {
			Discover("../internal/dynamicplugintesting", false)
			p := plugger.New("dynamicplugintesting").Plugins()
			Expect(p).To(BeEmpty())
		})

		It("discovers so test plugin in subdir", func() {
			Discover("../internal/dynamicplugintesting", true)
			p := plugger.New("dynamicplugintesting").Plugins()
			Expect(p).To(HaveLen(1))
			Expect(p[0].Name).To(Equal("dynfoo"))
		})

		It("finds the exported named function", func() {
			f := plugger.New("dynamicplugintesting").Func("PlugFunc")
			Expect(f).To(HaveLen(1))
			Expect(f[0].(func() string)()).To(Equal("dynfooplug"))
		})

	})

	Describe("plugin walking", func() {

		It("walks an existing plugin .so", func() {
			Expect(walkedOnSomething(
				false, "../internal/dynamicplugintesting/dynfoo/dynfooplug.so",
				mockedFileInfo{name: "dynfooplug.so", isdir: false},
				nil)).To(Succeed())
		})

		It("skips something else than .so", func() {
			Expect(walkedOnSomething(
				false, "plugins/foo/foo.bar",
				mockedFileInfo{name: "foo.bar", isdir: false},
				nil)).To(Succeed())
		})

		It("wants to walk into sub directories", func() {
			Expect(walkedOnSomething(
				false, "plugins/foo",
				mockedFileInfo{name: "foo", isdir: true},
				nil)).To(Equal(filepath.SkipDir))
		})

	})

})
