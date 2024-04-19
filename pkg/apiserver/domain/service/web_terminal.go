package service

import (
	"context"
	"errors"
	"fmt"
	"kdp-oam-operator/pkg/apiserver/domain/entity"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/controllers/bdc/constants"
	"kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/pkg/utils/log"
	"strings"
	"time"

	csv1alpha1 "github.com/cloudtty/cloudtty/pkg/apis/cloudshell/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WebTerminalService Terminal Service
type WebTerminalService interface {
	CreateTerminal(ctx context.Context, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName string) error
	GetExecTerminal(ctx context.Context, TerminalName, TerminalNameSpace string) (*entity.WebTerminalEntity, error)
	CheckTerminal(ctx context.Context, TerminalName, TerminalNameSpace string) error
	OpenTerminal(ctx context.Context, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName string) (*entity.WebTerminalEntity, error)
}

type webTerminalServiceImpl struct {
	KubeClient client.Client
	KubeConfig *rest.Config
}

// NewWebTerminalService new web terminal service
func NewWebTerminalService() WebTerminalService {
	kubeConfig, err := clients.GetKubeConfig()
	if err != nil {
		log.Logger.Fatalf("get kube config failure %s", err.Error())
	}
	kubeClient, err := clients.GetKubeClient()
	if err != nil {
		log.Logger.Fatalf("get kube client failure %s", err.Error())
	}
	return &webTerminalServiceImpl{
		KubeClient: kubeClient,
		KubeConfig: kubeConfig,
	}
}

func (w webTerminalServiceImpl) CreateTerminal(ctx context.Context, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName string) error {
	var command string
	ttl := utils.GetEnv("TTL", "3600")
	if podName != "" {
		command = fmt.Sprintf("kubectl exec -it %s -n %s -c %s -- sh -c \"clear; (bash || ash || sh)\"", podName, podNameSpace, containerName)
	} else {
		command = "bash"
	}

	cloudShell := csv1alpha1.CloudShell{
		TypeMeta: metav1.TypeMeta{
			Kind:       kindTerminal,
			APIVersion: kindTerminalApiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TerminalName,
			Namespace: TerminalNameSpace,
			Labels: map[string]string{
				constants.LabelName: TerminalName,
			},
		},
		Spec: csv1alpha1.CloudShellSpec{
			SecretRef: &csv1alpha1.LocalSecretReference{
				Name: kubeConfigSecretName,
			},
			CommandAction: command,
			Ttl:           utils.StringToInt32(ttl, 3600),
			Cleanup:       true,
			ExposeMode:    csv1alpha1.ExposureMode(utils.GetEnv("EXPOSURE_MODE", "ClusterIP")),
		},
	}
	if err := w.KubeClient.Create(ctx, &cloudShell); err != nil {
		return err
	}
	return nil
}

func (w webTerminalServiceImpl) GetExecTerminal(ctx context.Context, TerminalName, TerminalNameSpace string) (*entity.WebTerminalEntity, error) {
	err := w.CheckTerminal(ctx, TerminalName, TerminalNameSpace)
	if err == nil {
		return nil, errors.New("terminal not found")
	}

	var execTerminal csv1alpha1.CloudShell
	if err := w.KubeClient.Get(ctx, client.ObjectKey{Name: TerminalName, Namespace: TerminalNameSpace}, &execTerminal); err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Logger.Errorf("get %s %s status:%s", TerminalNameSpace, TerminalName, err.Error())
			return nil, errors.New("terminal not found")
		}
		return nil, err
	}
	if execTerminal.Status.Phase != "Ready" {
		log.Logger.Infof("get %s %s status:%s", TerminalNameSpace, TerminalName, execTerminal.Status.Phase)
		time.Sleep(500 * time.Millisecond)
		return w.GetExecTerminal(ctx, TerminalName, TerminalNameSpace)
	}
	return entity.Object2WebTerminalEntity(&execTerminal), nil
}

func (w webTerminalServiceImpl) CheckTerminal(ctx context.Context, TerminalName, TerminalNameSpace string) error {
	var execTerminal csv1alpha1.CloudShell
	if err := w.KubeClient.Get(ctx, client.ObjectKey{Name: TerminalName, Namespace: TerminalNameSpace}, &execTerminal); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	return errors.New("terminal is exists")
}

func (w webTerminalServiceImpl) OpenTerminal(ctx context.Context, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName string) (*entity.WebTerminalEntity, error) {

	//check pod exec
	err := w.CheckTerminal(ctx, TerminalName, TerminalNameSpace)
	if err != nil {
		if err.Error() == "terminal is exists" {
			log.Logger.Infof("%s %s %s", TerminalNameSpace, TerminalName, err.Error())
		} else {
			log.Logger.Errorf("check general exec failure %s", err.Error())
			return nil, err
		}

	} else {
		// create pod exec cloud shell
		err := w.CreateTerminal(ctx, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName)
		if err != nil {
			log.Logger.Errorf("create terminal exec failure %s", err.Error())
			return nil, errors.New(fmt.Sprintf("create terminal failed %s", err.Error()))
		}
		log.Logger.Infof("%s %s create success", TerminalNameSpace, TerminalName)
		time.Sleep(500 * time.Millisecond)
	}

	// get cloud shell status and url
	terminal, err := w.GetExecTerminal(ctx, TerminalName, TerminalNameSpace)
	if err != nil {
		return nil, nil
	}
	return terminal, nil
}
