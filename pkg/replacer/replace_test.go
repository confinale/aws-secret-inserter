package replacer

import (
	"os"
	"testing"
)

func Test_replaceSecrets(t *testing.T) {
	err := os.Setenv("_SUPERDUPER_LONG_ERRRORO_VAR", "123")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	type args struct {
		in string
		r  replacer
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		"single",
		args{
			in: "XXX = ::SECRET:x1:SECRET::",
			r: func(in string) string {
				if in == "x1" {
					return "X1"
				}
				return "not found"
			},
		},
		"XXX = X1",
	}, {
		"single-var",
		args{
			in: "XXX = ::SECRET:x1-${_SUPERDUPER_LONG_ERRRORO_VAR}:SECRET::",
			r: func(in string) string {
				if in == "x1-123" {
					return "X1"
				}
				return "not found"
			},
		},
		"XXX = X1",
	}, {
		"multi",
		args{
			in: "XXX = ::SECRET:x1:SECRET::\n\n::SECRET:x1:SECRET::\n::SEC:RET:x1:SECRET::\n::SECRET:x2:SECRET::",
			r: func(in string) string {
				if in == "x1" {
					return "X1"
				}
				return "not found"
			},
		},
		"XXX = X1\n\nX1\n::SEC:RET:x1:SECRET::\nnot found",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceSecrets(tt.args.in, tt.args.r)
			if got != tt.want {
				t.Errorf("replaceSecrets() got = %v, want %v", got, tt.want)
			}
		})
	}
}
