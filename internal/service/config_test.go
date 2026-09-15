package service

import "testing"

func TestConfigFromArgsDefaults(t *testing.T) {
	config, err := ConfigFromArgs([]string{"--database-url", "postgresql://localhost/myhealth"})
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != defaultHost || config.Port != defaultPort {
		t.Fatalf("unexpected address config: %+v", config)
	}
	if config.DatabaseURL != "postgresql://localhost/myhealth" || config.RequireClientCertificate {
		t.Fatalf("unexpected client config: %+v", config)
	}
}

func TestConfigFromArgsValues(t *testing.T) {
	config, err := ConfigFromArgs([]string{
		"--host", "::1", "--port", "8080",
		"--database-url", "postgresql://db/myhealth",
		"--require-client-cert",
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "::1" || config.Port != 8080 {
		t.Fatalf("unexpected address config: %+v", config)
	}
	if config.DatabaseURL != "postgresql://db/myhealth" || !config.RequireClientCertificate {
		t.Fatalf("unexpected client config: %+v", config)
	}
}

func TestConfigFromArgsRejectsInvalidPort(t *testing.T) {
	for _, port := range []string{"invalid", "0", "65536"} {
		t.Run(port, func(t *testing.T) {
			if _, err := ConfigFromArgs([]string{"--database-url", "postgresql://localhost/myhealth", "--port", port}); err == nil {
				t.Fatalf("--port=%q must be rejected", port)
			}
		})
	}
}

func TestConfigFromArgsRequiresDatabaseURL(t *testing.T) {
	if _, err := ConfigFromArgs(nil); err == nil {
		t.Fatal("missing --database-url must be rejected")
	}
}
