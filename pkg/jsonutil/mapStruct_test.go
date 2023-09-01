package jsonutil

import (
	"encoding/json"
	"fmt"
	"testing"
)

type SubStruct02 struct {
	Bool01 bool
}

func (s *SubStruct02) String() string {
	return fmt.Sprintf("%v", s.Bool01)
}

type Substruct01 struct {
	Double01 float64
}

func (s *Substruct01) String() string {
	return fmt.Sprintf("%v", s.Double01)
}

type TestStruct struct {
	Overflow `json:"-"`
	Str      string  `json:"string01,omitempty"`
	Str2     string  `json:"string02"`
	Str3     *string `json:"string03,omitempty"`
	Struct01 Substruct01
	Struct02 *SubStruct02 `json:"struct02,omitempty"`
	Int      int
	Int2     *int64
}

func (s *TestStruct) MarshalJSON() ([]byte, error) {
	return MarshalStructWithMap(s)
}

func (s *TestStruct) UnmarshalJSON(data []byte) error {
	return UnmarshalStructWithMap(data, s)
}

var int2 int64 = 48
var str3 string = ""
var ov1 = JSONBytes("62")
var test1 = &TestStruct{
	//	Overflow: Overflow{"ov01": 62, "ov02": "hello"},
	Overflow: Overflow{"ov01": &ov1, "ov2": "testing 123"},
	Str:      "test",
	Int:      42,
	Int2:     &int2,
	Str3:     &str3,
	Struct01: struct{ Double01 float64 }{Double01: 0.42},
	Struct02: &SubStruct02{Bool01: true},
}

func TestMapStruct(t *testing.T) {
	data, err := json.MarshalIndent(test1, "", "  ")
	if err != nil {
		t.Errorf("cannot marshal test1: %v", err)
	}
	fmt.Println(string(data))

	var test2 = &TestStruct{}
	if err := json.Unmarshal(data, test2); err != nil {
		t.Errorf("cannot unmarshal test1 data: %v", err)
	}
	fmt.Println(test2)
}
