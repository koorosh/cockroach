// Code generated by "stringer"; DO NOT EDIT.

package cli

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[raw-0]
	_ = x[human-1]
	_ = x[rangeID-2]
	_ = x[hex-3]
}

const _keyType_name = "rawhumanrangeIDhex"

var _keyType_index = [...]uint8{0, 3, 8, 15, 18}

func (i keyType) String() string {
	if i < 0 || i >= keyType(len(_keyType_index)-1) {
		return "keyType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _keyType_name[_keyType_index[i]:_keyType_index[i+1]]
}
