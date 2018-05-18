package main

import (
	"os"
	"testing"
)

func TestProto(t *testing.T) {
	os.Setenv("PCP_HUBUSERPASSWORD", "example")
	runProtoform("protoform.json")
}
