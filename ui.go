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

package steven

import "github.com/thinkofdeath/steven/ui"

func newButtonText(str string, x, y, w, h float64) (*ui.Button, *ui.Text) {
	btn := ui.NewButton(x, y, w, h)
	text := ui.NewText(str, 0, 0, 255, 255, 255).Attach(ui.Middle, ui.Center)
	text.Parent = btn
	btn.HoverFunc = func(over bool) {
		if over && !btn.Disabled {
			text.B = 160
		} else {
			text.B = 255
		}
	}
	return btn, text
}
