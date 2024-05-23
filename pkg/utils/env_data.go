package utils

func GetTTL() string {
	return GetEnv("TTL", "3600")
}

func GetHTTPType() string {
	return GetEnv("HTTPTYPE", "http")
}

func GetDOMAIN() string {
	return GetEnv("DOMAIN", "cloudtty.kdp-e2e.io")
}

func GetIngressName() string {
	return GetEnv("INGRESSNAME", "cloudtty")
}

func GetIngressClassName() string {
	return GetEnv("INGRESSCLASSNAME", "kong")
}

func GetTerminalTemplateName() string {
	return GetEnv("TERMINAL_TEMPLATE_NAME", "/opt/terminal-config/terminalTemplate.yaml")
}

func GetTerminalTransFormName() string {
	return GetEnv("TERMINAL_TRANSFORM_NAME", "/opt/terminal-config/terminalTransformData.json")
}

func GetMaxTry() string {
	return GetEnv("MAXTRY", "10")
}

func GetIngressTimeout() string {
	return GetEnv("INGRESSTIMEOUT", "0")
}
