package handler

import (
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	var (
		o                      = Options{}
		wantCmdReceiveDuration = time.Second
		wantAt                 = true
	)
	Apply([]SetOption{WithCmdReceiveDuration(wantCmdReceiveDuration), WithAt()},
		&o)

	if o.CmdReceiveDuration != wantCmdReceiveDuration {
		t.Errorf("unexpected CmdReceiveDuration, want %v actual %v",
			wantCmdReceiveDuration, o.CmdReceiveDuration)
	}

	if o.At != wantAt {
		t.Errorf("unexpected wantAt, want %v actual %v", wantAt, o.At)
	}

}
