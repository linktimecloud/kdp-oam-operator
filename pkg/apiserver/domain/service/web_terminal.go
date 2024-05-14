package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	v1dto "kdp-oam-operator/pkg/apiserver/apis/v1/dto"
	"kdp-oam-operator/pkg/apiserver/infrastructure/clients"
	"kdp-oam-operator/pkg/utils"
	"kdp-oam-operator/pkg/utils/log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"text/template"
	"time"

	"k8s.io/client-go/rest"
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
	ttl := utils.GetTTL()
	ingressName := utils.GetIngressName()
	ingressClassName := utils.GetIngressClassName()
	if podName != "" {
		command = fmt.Sprintf("kubectl exec -it %s -n %s -c %s -- sh -c \" (bash || ash || sh)\"", podName, podNameSpace, containerName)
	} else {
		command = "bash"
	}

	// Get cloud shell template file and parse
	terminalTemplateFileName := utils.GetTerminalTemplateName()
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
	MaxTry := utils.StringToInt(utils.GetMaxTry(), 10)
	err := w.CheckTerminal(ctx, TerminalName, TerminalNameSpace)
	if err == nil {
		return nil, errors.New("terminalNotFound")
	}

	// create obj to save data
	obj := &unstructured.Unstructured{}
	obj.SetKind("CloudShell")
	obj.SetAPIVersion("cloudshell.cloudtty.io/v1alpha1")
	if err := w.KubeClient.Get(ctx, client.ObjectKey{Name: TerminalName, Namespace: TerminalNameSpace}, obj); err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Logger.Errorf("get %s %s status:%s", TerminalNameSpace, TerminalName, err.Error())
			return nil, errors.New("terminalNotFound")
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
			return nil, errors.New("obtainLimitRetry")
		}
		log.Logger.Debugf("[%d/%d] get %s %s cloudshell status:%s", limitTry, MaxTry, TerminalNameSpace, TerminalName, terminalStatus)
		time.Sleep(500 * time.Millisecond)
		limitTry += 1
		return w.GetExecTerminal(ctx, TerminalName, TerminalNameSpace, limitTry)
	}
	log.Logger.Infof("get %s %s cloudshell status:%s", TerminalNameSpace, TerminalName, terminalStatus)
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
			return nil, errors.New("createTerminalFailed")
		}
		log.Logger.Infof("create %s %s terminal success", TerminalNameSpace, TerminalName)
		time.Sleep(1 * time.Second)
	}

	// get terminal
	terminal, err := w.GetExecTerminal(ctx, TerminalName, TerminalNameSpace, 0)
	if err != nil {
		return nil, err
	}

	// deal with ingress route not match
	ingressTimeout := utils.StringToInt64(utils.GetIngressTimeout(), 0)
	time.Sleep(time.Second * time.Duration(ingressTimeout))
	return terminal, nil
}

// ExtractData Extract the data according to the rules
func ExtractData(obj *unstructured.Unstructured, rules v1dto.ExtractionRules) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	// Iterate over the extraction rule
	for key, path := range map[string][]string{
		"phase":     rules.Phase,
		"accessUrl": rules.AccessUrl,
		"ttl":       rules.Ttl,
	} {
		// According to the path to access unstructured. Unstructured object
		value, found, err := unstructured.NestedFieldNoCopy(obj.Object, path...)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("key '%s' not found", key)
		}
		result[key] = value
	}

	return result, nil
}

// GetTerminalData get terminal url from terminal response
func GetTerminalData(terminal *unstructured.Unstructured) (url, phase string, ttl int64, err error) {
	terminalUrl, terminalPhase, terminalTtl := "", "", int64(0)
	rules, err := ParseExtractionRules()
	if err != nil {
		fmt.Println("parse transform file data err:", err)
		return terminalUrl, terminalPhase, terminalTtl, err
	}
	data, err := ExtractData(terminal, rules)
	if err != nil {
		log.Logger.Errorf("Error extracting data: %s", err)
		return "", "", 0, err
	}
	// Assembling urls
	accessUrl := utils.GetStringValue(data, "accessUrl")
	terminalPhase = utils.GetStringValue(data, "phase")
	terminalTtl = utils.GetInt64Value(data, "ttl")
	return accessUrl, terminalPhase, terminalTtl, nil
}

func ParseExtractionRules() (v1dto.ExtractionRules, error) {
	terminalTransformFileName := utils.GetTerminalTransFormName()
	jsonFile, err := os.ReadFile(terminalTransformFileName)
	if err != nil {
		return v1dto.ExtractionRules{}, fmt.Errorf("error reading JSON file %s: %v", terminalTransformFileName, err)
	}

	var rules v1dto.ExtractionRules
	if err := json.Unmarshal(jsonFile, &rules); err != nil {
		return v1dto.ExtractionRules{}, fmt.Errorf("error parsing JSON: %v", err)
	}

	return rules, nil
}

func GetTerminalUrl(accessUrl string) string {
	httpType := utils.GetHTTPType()
	DOMAIN := utils.GetDOMAIN()
	terminalUrl := httpType + "://" + DOMAIN + accessUrl
	return terminalUrl
}
