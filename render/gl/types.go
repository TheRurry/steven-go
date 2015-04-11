// Copyright 2015 Matthew Collins
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

package gl

import (
	"github.com/thinkofdeath/gl/v3.2-core/gl"
)

// Type is a type of data used by various operations.
type Type uint32

// Valid types.
const (
	UnsignedByte  Type = gl.UNSIGNED_BYTE
	UnsignedShort Type = gl.UNSIGNED_SHORT
	Short         Type = gl.SHORT
	Float         Type = gl.FLOAT
)