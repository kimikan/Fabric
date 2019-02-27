package handler_test

import (
	"FabricSdkDemo/handler"
	"testing"
)

func TestJson(t *testing.T) {
	t.Error(handler.JsonToStrings("[\"1\"]"))
}
