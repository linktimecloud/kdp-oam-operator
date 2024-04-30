package utils

import (
	"os"
	"testing"
)

// TestGetTTL tests the GetTTL function
func TestGetTTL(t *testing.T) {
	// Set up environment variable for testing
	os.Setenv("TTL", "7200")
	defer os.Unsetenv("TTL") // Clean up after test

	// Test with the environment variable set
	expected := "7200"
	got := GetTTL()
	if got != expected {
		t.Errorf("GetTTL() = %v; want %v", got, expected)
	}

	// Test with no environment variable set (should return default value)
	os.Unsetenv("TTL")
	expected = "3600" // Default value
	got = GetTTL()
	if got != expected {
		t.Errorf("GetTTL() = %v; want %v", got, expected)
	}
}

// TestGetHTTPType tests the GetHTTPType function
func TestGetHTTPType(t *testing.T) {
	// Set up environment variable for testing
	os.Setenv("HTTPTYPE", "https")
	defer os.Unsetenv("HTTPTYPE") // Clean up after test

	// Test with the environment variable set
	expected := "https"
	got := GetHTTPType()
	if got != expected {
		t.Errorf("GetHTTPType() = %v; want %v", got, expected)
	}

	// Test with no environment variable set (should return default value)
	os.Unsetenv("HTTPTYPE")
	expected = "http" // Default value
	got = GetHTTPType()
	if got != expected {
		t.Errorf("GetHTTPType() = %v; want %v", got, expected)
	}
}

// TestGetDOMAIN tests the GetDOMAIN function for default value and environment override
func TestGetDOMAIN(t *testing.T) {
	const defaultDomain = "cloudtty.kdp-e2e.io"
	const testDomain = "test.example.com"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("DOMAIN") // Ensure env var is not set
		defer os.Unsetenv("DOMAIN")

		got := GetDOMAIN()
		if got != defaultDomain {
			t.Errorf("GetDOMAIN() = %v; want %v", got, defaultDomain)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("DOMAIN", testDomain)
		defer os.Unsetenv("DOMAIN")

		got := GetDOMAIN()
		if got != testDomain {
			t.Errorf("GetDOMAIN() = %v; want %v", got, testDomain)
		}
	})
}

// TestGetIngressName tests the GetIngressName function for default value and environment override
func TestGetIngressName(t *testing.T) {
	const defaultIngressName = "cloudtty"
	const testIngressName = "test-ingress"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("INGRESSNAME") // Ensure env var is not set
		defer os.Unsetenv("INGRESSNAME")

		got := GetIngressName()
		if got != defaultIngressName {
			t.Errorf("GetIngressName() = %v; want %v", got, defaultIngressName)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("INGRESSNAME", testIngressName)
		defer os.Unsetenv("INGRESSNAME")

		got := GetIngressName()
		if got != testIngressName {
			t.Errorf("GetIngressName() = %v; want %v", got, testIngressName)
		}
	})
}

// TestGetIngressClassName tests the GetIngressClassName function for default value and environment override
func TestGetIngressClassName(t *testing.T) {
	const defaultIngressClassName = "kong"
	const testIngressClassName = "nginx"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("INGRESSCLASSNAME") // Ensure env var is not set
		defer os.Unsetenv("INGRESSCLASSNAME")

		got := GetIngressClassName()
		if got != defaultIngressClassName {
			t.Errorf("GetIngressClassName() = %v; want %v", got, defaultIngressClassName)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("INGRESSCLASSNAME", testIngressClassName)
		defer os.Unsetenv("INGRESSCLASSNAME")

		got := GetIngressClassName()
		if got != testIngressClassName {
			t.Errorf("GetIngressClassName() = %v; want %v", got, testIngressClassName)
		}
	})
}

// TestGetTerminalTemplateName tests the GetTerminalTemplateName function for default value and environment override
func TestGetTerminalTemplateName(t *testing.T) {
	const defaultTerminalTemplateName = "/opt/terminal-config/terminalTemplate.yaml"
	const testTerminalTemplateName = "/custom/path/to/template.yaml"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("TERMINAL_TEMPLATE_NAME") // Ensure env var is not set
		defer os.Unsetenv("TERMINAL_TEMPLATE_NAME")

		got := GetTerminalTemplateName()
		if got != defaultTerminalTemplateName {
			t.Errorf("GetTerminalTemplateName() = %v; want %v", got, defaultTerminalTemplateName)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("TERMINAL_TEMPLATE_NAME", testTerminalTemplateName)
		defer os.Unsetenv("TERMINAL_TEMPLATE_NAME")

		got := GetTerminalTemplateName()
		if got != testTerminalTemplateName {
			t.Errorf("GetTerminalTemplateName() = %v; want %v", got, testTerminalTemplateName)
		}
	})
}

func TestGetTerminalTransFormName(t *testing.T) {
	const defaultTransformName = "/opt/terminal-config/terminalTransformData.json"
	const testTransformName = "/custom/path/to/transformData.json"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("TERMINAL_TRANSFORM_NAME")
		defer os.Unsetenv("TERMINAL_TRANSFORM_NAME")

		got := GetTerminalTransFormName()
		if got != defaultTransformName {
			t.Errorf("GetTerminalTransFormName() = %v; want %v", got, defaultTransformName)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("TERMINAL_TRANSFORM_NAME", testTransformName)
		defer os.Unsetenv("TERMINAL_TRANSFORM_NAME")

		got := GetTerminalTransFormName()
		if got != testTransformName {
			t.Errorf("GetTerminalTransFormName() = %v; want %v", got, testTransformName)
		}
	})
}

func TestGetMaxTry(t *testing.T) {
	const defaultMaxTry = "10"
	const testMaxTry = "5"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("MAXTRY")
		defer os.Unsetenv("MAXTRY")

		got := GetMaxTry()
		if got != defaultMaxTry {
			t.Errorf("GetMaxTry() = %v; want %v", got, defaultMaxTry)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("MAXTRY", testMaxTry)
		defer os.Unsetenv("MAXTRY")

		got := GetMaxTry()
		if got != testMaxTry {
			t.Errorf("GetMaxTry() = %v; want %v", got, testMaxTry)
		}
	})
}

func TestGetIngressTimeout(t *testing.T) {
	const defaultIngressTimeout = "0"
	const testIngressTimeout = "30"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("INGRESSTIMEOUT")
		defer os.Unsetenv("INGRESSTIMEOUT")

		got := GetIngressTimeout()
		if got != defaultIngressTimeout {
			t.Errorf("GetIngressTimeout() = %v; want %v", got, defaultIngressTimeout)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("INGRESSTIMEOUT", testIngressTimeout)
		defer os.Unsetenv("INGRESSTIMEOUT")

		got := GetIngressTimeout()
		if got != testIngressTimeout {
			t.Errorf("GetIngressTimeout() = %v; want %v", got, testIngressTimeout)
		}
	})
}

func TestGetProxyHost(t *testing.T) {
	const defaultProxyHost = "0"
	const testProxyHost = "proxy.example.com"

	t.Run("Default Value", func(t *testing.T) {
		os.Unsetenv("KONG_KONG_PROXY_SERVICE_HOST")
		defer os.Unsetenv("KONG_KONG_PROXY_SERVICE_HOST")

		got := GetProxyHost()
		if got != defaultProxyHost {
			t.Errorf("GetProxyHost() = %v; want %v", got, defaultProxyHost)
		}
	})

	t.Run("Environment Override", func(t *testing.T) {
		os.Setenv("KONG_KONG_PROXY_SERVICE_HOST", testProxyHost)
		defer os.Unsetenv("KONG_KONG_PROXY_SERVICE_HOST")

		got := GetProxyHost()
		if got != testProxyHost {
			t.Errorf("GetProxyHost() = %v; want %v", got, testProxyHost)
		}
	})
}
