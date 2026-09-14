package service

import "testing"

func TestConfigFromEnvironmentDefaults(t *testing.T) {
	t.Setenv("HOST", "")
	t.Setenv("PORT", "")
	t.Setenv("CLIENT_ORIGIN", "")
	t.Setenv("REQUIRE_CLIENT_CERT", "")

	config, err := ConfigFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != defaultHost || config.Port != defaultPort {
		t.Fatalf("unexpected address config: %+v", config)
	}
	if config.ClientOrigin != defaultClientOrigin || config.RequireClientCertificate {
		t.Fatalf("unexpected client config: %+v", config)
	}
}

func TestConfigFromEnvironmentValues(t *testing.T) {
	t.Setenv("HOST", "::1")
	t.Setenv("PORT", "8080")
	t.Setenv("CLIENT_ORIGIN", "https://health.example.com")
	t.Setenv("REQUIRE_CLIENT_CERT", "true")

	config, err := ConfigFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "::1" || config.Port != 8080 {
		t.Fatalf("unexpected address config: %+v", config)
	}
	if config.ClientOrigin != "https://health.example.com" || !config.RequireClientCertificate {
		t.Fatalf("unexpected client config: %+v", config)
	}
}

func TestConfigFromEnvironmentRejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"invalid", "0", "65536"} {
		t.Run(port, func(t *testing.T) {
			t.Setenv("PORT", port)
			if _, err := ConfigFromEnvironment(); err == nil {
				t.Fatalf("PORT=%q must be rejected", port)
			}
		})
	}
}
