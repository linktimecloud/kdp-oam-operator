package common

import (
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestObjectReference_Equal(t *testing.T) {
	type fields struct {
		ObjectReference v1.ObjectReference
	}
	type args struct {
		r ObjectReference
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Equal references",
			fields: fields{
				ObjectReference: v1.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "example-pod",
					Namespace:  "default",
				},
			},
			args: args{
				r: ObjectReference{
					ObjectReference: v1.ObjectReference{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "example-pod",
						Namespace:  "default",
					},
				},
			},
			want: true,
		},
		{
			name: "Different APIVersion",
			fields: fields{
				ObjectReference: v1.ObjectReference{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "example-pod1",
					Namespace:  "default",
				},
			},
			args: args{
				r: ObjectReference{
					ObjectReference: v1.ObjectReference{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "example-pod",
						Namespace:  "default",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ObjectReference{
				ObjectReference: tt.fields.ObjectReference,
			}
			if got := in.Equal(tt.args.r); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
