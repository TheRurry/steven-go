// generated by stringer -type=Action; DO NOT EDIT

package platform

import "fmt"

const _Action_name = "InvalidActionDebug"

var _Action_index = [...]uint8{0, 13, 18}

func (i Action) String() string {
	if i < 0 || i+1 >= Action(len(_Action_index)) {
		return fmt.Sprintf("Action(%d)", i)
	}
	return _Action_name[_Action_index[i]:_Action_index[i+1]]
}