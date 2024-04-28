package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/pkg/utils/log"
	"strings"
	"text/template"
	"time"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WebTerminalService Terminal Service
type WebTerminalService interface {
	CreateTerminal(ctx context.Context, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName string) error
	GetExecTerminal(ctx context.Context, TerminalName, TerminalNameSpace string, limitTry int) (*unstructured.Unstructured, error)
	CheckTerminal(ctx context.Context, TerminalName, TerminalNameSpace string) error
	OpenTerminal(ctx context.Context, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName string) (*unstructured.Unstructured, error)
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
	ingressName := utils.GetEnv("INGRESSNAME", "cloudtty")
	ingressClassName := utils.GetEnv("INGRESSCLASSNAME", "kong")
	if podName != "" {
		command = fmt.Sprintf("kubectl exec -it %s -n %s -c %s -- sh -c \"clear; (bash || ash || sh)\"", podName, podNameSpace, containerName)
	} else {
		command = "bash"
	}

	// Get cloud shell template file and parse
	terminalTemplateFileName := utils.GetEnv("TERMINAL_TEMPLATE_NAME", "/opt/terminal-config/terminalTemplate.yaml")
	tmpl, err := template.ParseFiles(terminalTemplateFileName)
	if err != nil {
		log.Logger.Errorf("parse template file %s status:%s", terminalTemplateFileName, err.Error())
		return err
	}

	data := struct {
		TerminalName           string
		CommandAction          string
		TerminalNameSpace      string
		TtlSecondsAfterStarted int64
		KubeConfigName         string
		IngressName            string
		IngressClassName       string
	}{
		TerminalName:           TerminalName,
		CommandAction:          command,
		TerminalNameSpace:      TerminalNameSpace,
		TtlSecondsAfterStarted: utils.StringToInt64(ttl, 3600),
		KubeConfigName:         kubeConfigSecretName,
		IngressName:            ingressName,
		IngressClassName:       ingressClassName,
	}

	// Create template and render
	var RenderTerminalData bytes.Buffer
	err = tmpl.Execute(&RenderTerminalData, data)
	if err != nil {
		log.Logger.Errorf("create template and render status:%s", err.Error())
		return err
	}

	// Converts the rendered YAML string into a Kubernetes resource object
	var obj unstructured.Unstructured
	decode := yaml.NewYAMLOrJSONDecoder(&RenderTerminalData, 4096)
	err = decode.Decode(&obj.Object)
	if err != nil {
		log.Logger.Errorf("render yaml to Kubernetes resource object status:%s", err.Error())
		return err
	}

	if err := w.KubeClient.Create(ctx, &obj); err != nil {
		log.Logger.Errorf("create %s %s terminal status:%s", TerminalNameSpace, TerminalName, err.Error())
		return err
	}
	return nil
}

func (w webTerminalServiceImpl) GetExecTerminal(ctx context.Context, TerminalName, TerminalNameSpace string, limitTry int) (*unstructured.Unstructured, error) {
	MaxTry := utils.StringToInt(utils.GetEnv("MAXTRY", "10"), 10)
	err := w.CheckTerminal(ctx, TerminalName, TerminalNameSpace)
	if err == nil {
		return nil, errors.New("terminal not found")
	}

	// create obj to save data
	obj := &unstructured.Unstructured{}
	obj.SetKind("CloudShell")
	obj.SetAPIVersion("cloudshell.cloudtty.io/v1alpha1")
	if err := w.KubeClient.Get(ctx, client.ObjectKey{Name: TerminalName, Namespace: TerminalNameSpace}, obj); err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Logger.Errorf("get %s %s status:%s", TerminalNameSpace, TerminalName, err.Error())
			return nil, errors.New("terminal not found")
		}
		return nil, err
	}

	terminalStatus, found, _ := unstructured.NestedFieldCopy(obj.Object, "status", "phase")
	if !found {
		fmt.Println("Status:", terminalStatus)
	}

	if terminalStatus != "Ready" {
		if limitTry > MaxTry {
			log.Logger.Errorf("%s %s get status The number of retries exceeded the maximum，status:%s", TerminalNameSpace, TerminalName, terminalStatus)
			return nil, errors.New(fmt.Sprintf("The number of attempts to obtain terminal status exceeded the maximum， status is:%s", terminalStatus))
		}
		log.Logger.Infof("[%d/%d] get %s %s cloudshell status:%s", limitTry, MaxTry, TerminalNameSpace, TerminalName, terminalStatus)
		time.Sleep(500 * time.Millisecond)
		limitTry += 1
		return w.GetExecTerminal(ctx, TerminalName, TerminalNameSpace, limitTry)
	}
	log.Logger.Infof("%s %s get data success", TerminalNameSpace, TerminalName)
	return obj, nil
}

func (w webTerminalServiceImpl) CheckTerminal(ctx context.Context, TerminalName, TerminalNameSpace string) error {
	obj := &unstructured.Unstructured{}
	obj.SetKind("CloudShell")
	obj.SetAPIVersion("cloudshell.cloudtty.io/v1alpha1")
	if err := w.KubeClient.Get(ctx, client.ObjectKey{Name: TerminalName, Namespace: TerminalNameSpace}, obj); err != nil {
		log.Logger.Infof("get %s %s status:%s", TerminalNameSpace, TerminalName, err.Error())
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return err
	}
	return errors.New("terminal is exists")
}

func (w webTerminalServiceImpl) OpenTerminal(ctx context.Context, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName string) (*unstructured.Unstructured, error) {
	//check terminal
	err := w.CheckTerminal(ctx, TerminalName, TerminalNameSpace)
	if err != nil {
		if err.Error() == "terminal is exists" {
			log.Logger.Infof("%s %s %s", TerminalNameSpace, TerminalName, err.Error())
		} else {
			log.Logger.Errorf("check general exec failure %s", err.Error())
			return nil, err
		}

	} else {
		// create terminal
		err := w.CreateTerminal(ctx, kubeConfigSecretName, TerminalName, TerminalNameSpace, podNameSpace, podName, containerName)
		if err != nil {
			log.Logger.Errorf("create terminal exec failure %s", err.Error())
			return nil, errors.New(fmt.Sprintf("create terminal failed %s", err.Error()))
		}
		log.Logger.Infof("%s %s create success", TerminalNameSpace, TerminalName)
		time.Sleep(500 * time.Millisecond)
	}

	// get terminal and url
	terminal, err := w.GetExecTerminal(ctx, TerminalName, TerminalNameSpace, 0)
	if err != nil {
		return nil, err
	}
	return terminal, nil
}
