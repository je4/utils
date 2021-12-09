package mimetypes

import (
	"testing"
)

var m2e = NewMimeExt()

func TestMime2Ext(t *testing.T) {
	strE2M := m2e.SPrintExt2Mime()
	strM2E := m2e.SPrintMime2Ext()

	if len(strE2M) == 0 {
		t.Fatalf("no E2M")
	}
	if len(strM2E) == 0 {
		t.Fatalf("no M2E")
	}
	println(strM2E)
}
