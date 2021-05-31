// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package term

// Compiler interface represents term compilation algorithm
type Compiler interface {
	Compile(interface{}) (termValue, compiledValue interface{}, err error)
}

type emptyCompiler struct{}

// Compile returns the value unmodified
func (emptyCompiler) Compile(value interface{}) (termValue, compiledValue interface{}, err error) {
	termValue = value
	return
}

// GetCompiler return a compiler that will compile the value
func GetCompiler(meta interface{}) Compiler {
	c, ok := meta.(Compiler)
	if ok {
		return c
	}
	return emptyCompiler{}
}
