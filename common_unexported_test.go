package dbw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_intersection(t *testing.T) {
	type args struct {
		av []string
		bv []string
	}
	tests := []struct {
		name       string
		args       args
		want       []string
		want1      map[string]string
		want2      map[string]string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "intersect",
			args: args{
				av: []string{"alice"},
				bv: []string{"alice", "bob"},
			},
			want: []string{"alice"},
			want1: map[string]string{
				"ALICE": "alice",
			},
			want2: map[string]string{
				"ALICE": "alice",
				"BOB":   "bob",
			},
		},
		{
			name: "intersect-2",
			args: args{
				av: []string{"alice", "bob", "jane", "doe"},
				bv: []string{"alice", "doe", "bert", "ernie", "bigbird"},
			},
			want: []string{"alice", "doe"},
			want1: map[string]string{
				"ALICE": "alice",
				"BOB":   "bob",
				"JANE":  "jane",
				"DOE":   "doe",
			},
			want2: map[string]string{
				"ALICE":   "alice",
				"DOE":     "doe",
				"BERT":    "bert",
				"ERNIE":   "ernie",
				"BIGBIRD": "bigbird",
			},
		},
		{
			name: "intersect-mixed-case",
			args: args{
				av: []string{"AlicE"},
				bv: []string{"alICe", "Bob"},
			},
			want: []string{"alice"},
			want1: map[string]string{
				"ALICE": "AlicE",
			},
			want2: map[string]string{
				"ALICE": "alICe",
				"BOB":   "Bob",
			},
		},
		{
			name: "no-intersect-mixed-case",
			args: args{
				av: []string{"AliCe", "BOb", "jaNe", "DOE"},
				bv: []string{"beRt", "ERnie", "bigBIRD"},
			},
			want: []string{},
			want1: map[string]string{
				"ALICE": "AliCe",
				"BOB":   "BOb",
				"JANE":  "jaNe",
				"DOE":   "DOE",
			},
			want2: map[string]string{
				"BERT":    "beRt",
				"ERNIE":   "ERnie",
				"BIGBIRD": "bigBIRD",
			},
		},
		{
			name: "no-intersect-1",
			args: args{
				av: []string{"alice", "bob", "jane", "doe"},
				bv: []string{"bert", "ernie", "bigbird"},
			},
			want: []string{},
			want1: map[string]string{
				"ALICE": "alice",
				"BOB":   "bob",
				"JANE":  "jane",
				"DOE":   "doe",
			},
			want2: map[string]string{
				"BERT":    "bert",
				"ERNIE":   "ernie",
				"BIGBIRD": "bigbird",
			},
		},
		{
			name: "empty-av",
			args: args{
				av: []string{},
				bv: []string{"bert", "ernie", "bigbird"},
			},
			want:  []string{},
			want1: map[string]string{},
			want2: map[string]string{
				"BERT":    "bert",
				"ERNIE":   "ernie",
				"BIGBIRD": "bigbird",
			},
		},
		{
			name: "empty-av-and-bv",
			args: args{
				av: []string{},
				bv: []string{},
			},
			want:  []string{},
			want1: map[string]string{},
			want2: map[string]string{},
		},
		{
			name: "nil-av",
			args: args{
				av: nil,
				bv: []string{"bert", "ernie", "bigbird"},
			},
			want:       nil,
			want1:      nil,
			want2:      nil,
			wantErr:    true,
			wantErrMsg: "dbw.Intersection: av is missing: invalid parameter",
		},
		{
			name: "nil-bv",
			args: args{
				av: []string{},
				bv: nil,
			},
			want:       nil,
			want1:      nil,
			want2:      nil,
			wantErr:    true,
			wantErrMsg: "dbw.Intersection: bv is missing: invalid parameter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			got, got1, got2, err := Intersection(tt.args.av, tt.args.bv)
			if err == nil && tt.wantErr {
				assert.Error(err)
			}
			if tt.wantErr {
				assert.Error(err)
				assert.Equal(tt.wantErrMsg, err.Error())
			}
			assert.Equal(tt.want, got)
			assert.Equal(tt.want1, got1)
			assert.Equal(tt.want2, got2)
		})
	}
}
