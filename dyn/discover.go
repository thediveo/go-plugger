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
	"plugin"
)

// Discover discovers plugins located at or within a specific path, optionally
// also (recursively) looking into subdirectories of path, and loads them, so
// the plugins can register themselves.
func Discover(path string, recursive bool) {
	// We handle also the non-recursive usecase with the ordinary filepath
	// walker, as this simplifies things enormously ... when combined with
	// closures.
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		return walkedOnSomething(recursive, path, info, err)
	})
}

//
// This is an example of when to separate out an enclosed callback function in
// order to allow testing it separatedly.
func walkedOnSomething(recursive bool, path string, info os.FileInfo, err error) error {
	if info != nil {
		if info.IsDir() {
			// If its a directory and we're not allowed to search
			// recursively for plugins, then tell the walker to please
			// stop here and to go elsewhere. Otherwise, let the walker
			// walk freely.
			if !recursive {
				return filepath.SkipDir
			}
		} else if filepath.Ext(info.Name()) == ".so" {
			// If it's a file and its name looks like a potential shared
			// library, then try to load it. If it fails, we keep silent,
			// because we want to look still for other plugins. Please note
			// that the loaded plugin is responsible to register itself.
			_, err = plugin.Open(path)
		}
	}
	return err
}
