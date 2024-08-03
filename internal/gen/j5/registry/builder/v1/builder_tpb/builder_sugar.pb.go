// Code generated by protoc-gen-go-sugar. DO NOT EDIT.

package builder_tpb

import (
	driver "database/sql/driver"
	fmt "fmt"
)

// BuildStatus
const (
	BuildStatus_UNSPECIFIED BuildStatus = 0
	BuildStatus_IN_PROGRESS BuildStatus = 1
	BuildStatus_SUCCESS     BuildStatus = 2
	BuildStatus_FAILURE     BuildStatus = 3
)

var (
	BuildStatus_name_short = map[int32]string{
		0: "UNSPECIFIED",
		1: "IN_PROGRESS",
		2: "SUCCESS",
		3: "FAILURE",
	}
	BuildStatus_value_short = map[string]int32{
		"UNSPECIFIED": 0,
		"IN_PROGRESS": 1,
		"SUCCESS":     2,
		"FAILURE":     3,
	}
	BuildStatus_value_either = map[string]int32{
		"UNSPECIFIED":              0,
		"BUILD_STATUS_UNSPECIFIED": 0,
		"IN_PROGRESS":              1,
		"BUILD_STATUS_IN_PROGRESS": 1,
		"SUCCESS":                  2,
		"BUILD_STATUS_SUCCESS":     2,
		"FAILURE":                  3,
		"BUILD_STATUS_FAILURE":     3,
	}
)

// ShortString returns the un-prefixed string representation of the enum value
func (x BuildStatus) ShortString() string {
	return BuildStatus_name_short[int32(x)]
}
func (x BuildStatus) Value() (driver.Value, error) {
	return []uint8(x.ShortString()), nil
}
func (x *BuildStatus) Scan(value interface{}) error {
	var strVal string
	switch vt := value.(type) {
	case []uint8:
		strVal = string(vt)
	case string:
		strVal = vt
	default:
		return fmt.Errorf("invalid type %T", value)
	}
	val := BuildStatus_value_either[strVal]
	*x = BuildStatus(val)
	return nil
}