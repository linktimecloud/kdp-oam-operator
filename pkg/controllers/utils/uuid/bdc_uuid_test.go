package uuid

import "testing"

func TestGenAppUUID(t *testing.T) {
	type args struct {
		namespace string
		name      string
		length    int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{namespace: "admin", name: "flink", length: 8}, "mtu5mzaw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenAppUUID(tt.args.namespace, tt.args.name, tt.args.length); got != tt.want {
				t.Errorf("GenAppUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genAppUUID(t *testing.T) {
	type args struct {
		hashStr string
		length  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{hashStr: "aaaaaaaaaaaa", length: 8}, "mjgynzy2"},
		{"test2", args{hashStr: "20d84cf2d4b85774a2399b8cee0f3131", length: 8}, "mtu5mzaw"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genAppUUID(tt.args.hashStr, tt.args.length); got != tt.want {
				t.Errorf("genAppUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_uuid5(t *testing.T) {
	type args struct {
		namespace string
		name      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test1", args{namespace: "admin", name: "flink"}, "20d84cf2d4b85774a2399b8cee0f3131"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uuid5(tt.args.namespace, tt.args.name); got != tt.want {
				t.Errorf("uuid5() = %v, want %v", got, tt.want)
			}
		})
	}
}
