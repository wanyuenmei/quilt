package elector

import (
	"github.com/NetSys/quilt/db"
	"testing"
)

func TestStartElector(t *testing.T) {
	conn := db.New()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Should have panicked")
		}
	}()
	commitLeader(conn, false, "first_ip", "second_ip")
}
