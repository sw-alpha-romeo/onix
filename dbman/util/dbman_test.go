package util

import "testing"

var dbman *DbMan

func init() {
	dbm, err := NewDbMan("", "")
	if err != nil {
		panic(err)
	}
	dbman = dbm
}

func TestFetchReleasePlan(t *testing.T) {
	plan, err := dbman.GetReleasePlan()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(plan.Releases) == 0 {
		t.Errorf("no releases found in plan")
		t.Fail()
	}
}

func TestSaveConfig(t *testing.T) {
	dbman.SetConfig("Schema.URI", "AAAA")
	dbman.SaveConfig()
}

func TestUseConfig(t *testing.T) {
	dbman.Use("", "myapp")
}
