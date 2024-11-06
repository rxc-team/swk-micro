package msg

import "testing"

func Test_objFormat(t *testing.T) {
	type args struct {
		temp   string
		params map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				temp: "{{.name}} 你好",
				params: map[string]string{
					"name": "小吴",
				},
			},
			want: "小吴 你好",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := objFormat(tt.args.temp, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("objFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("objFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
