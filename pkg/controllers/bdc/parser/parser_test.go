package parser

import (
	"context"
	bdcv1alpha1 "kdp-oam-operator/api/bdc/v1alpha1"
	pkgcommon "kdp-oam-operator/pkg/common"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/controllers/bdc/defcontext"
	"kdp-oam-operator/pkg/controllers/utils/uuid"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestGenerateContextDataFromBigDataClusterFile(t *testing.T) {
	cli := fake.NewClientBuilder().Build()

	req := ctrl.Request{
		NamespacedName: client.ObjectKey{
			Name:      "test-app",
			Namespace: "test-ns",
		},
	}

	wantData := defcontext.ContextData{
		Name:      "test-app",
		BDCName:   "test-bdc",
		Namespace: "test-ns",
		BDCAnnotations: map[string]string{
			constants.AnnotationBDCName: "test-bdc",
			constants.AnnotationOrgName: "test-org",
		},
		BDCLabels: map[string]string{
			"test-label": "test-value",
		},
	}
	wantData.PushData("app_uuid", uuid.GenAppUUID(req.Namespace, req.Name, 8))
	wantData.PushData("bdc", "test-bdc")
	wantData.PushData("group", "test-org")

	type args struct {
		bdcFile *BDCFile
		ctx     context.Context
		req     ctrl.Request
	}
	tests := []struct {
		name    string
		args    args
		want    defcontext.ContextData
		wantErr bool
	}{
		{
			name: "test generate context data from big data cluster file",
			args: args{
				bdcFile: &BDCFile{
					BDCName:             "test-bdc",
					Name:                "test-app",
					DownStreamNamespace: "test-ns",
					BigDataClusterAnnotations: map[string]string{
						constants.AnnotationBDCName: "test-bdc",
						constants.AnnotationOrgName: "test-org",
					},
					BigDataClusterLabels: map[string]string{
						"test-label": "test-value",
					},
					Parser: NewParser(cli),
				},
				ctx: context.Background(),
				req: req,
			},
			want:    wantData,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateContextDataFromBigDataClusterFile(tt.args.bdcFile, tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateContextDataFromBigDataClusterFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateContextDataFromBigDataClusterFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateDownStreamNamespace(t *testing.T) {
	type args struct {
		bdc *bdcv1alpha1.BigDataCluster
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "default namespace",
			args: args{
				bdc: &bdcv1alpha1.BigDataCluster{
					Spec: bdcv1alpha1.BigDataClusterSpec{
						Namespaces: []bdcv1alpha1.Namespace{
							{Name: "test-ns", IsDefault: false},
							{Name: "default-ns", IsDefault: true},
						},
					},
				},
			},
			want:    "default-ns",
			wantErr: false,
		},
		{
			name: "not default namespace",
			args: args{
				bdc: &bdcv1alpha1.BigDataCluster{
					Spec: bdcv1alpha1.BigDataClusterSpec{
						Namespaces: []bdcv1alpha1.Namespace{
							{Name: "test-ns", IsDefault: false},
						},
					},
				},
			},
			want:    pkgcommon.SystemDefaultNamespace,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateDownStreamNamespace(tt.args.bdc)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDownStreamNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateDownStreamNamespace() got = %v, want %v", got, tt.want)
			}
		})
	}
}
