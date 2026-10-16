/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	sb_util "github.com/SENERGY-Platform/go-service-base/util"
)

var testSecrets = map[string]string{
	"KEYCLOAK_CLIENT_SECRET": "test-client-secret-1",
	"KEYCLOAK_PW":            "test-keycloak-password-2",
	"RANCHER2_ACCESS_KEY":    "test-access-key-3",
	"RANCHER2_SECRET_KEY":    "test-secret-key-4",
}

func assertMasked(t *testing.T, cfg *Config) {
	t.Helper()
	// same call as the start-log line in main.go
	out := sb_util.ToJsonStr(cfg)
	for name, value := range testSecrets {
		if strings.Contains(out, value) {
			t.Errorf("%s value is part of the logged config", name)
		}
	}
	if got := cfg.Keycloak.ClientSecret.Value(); got != testSecrets["KEYCLOAK_CLIENT_SECRET"] {
		t.Errorf("client secret not usable: %q", got)
	}
	if got := cfg.Keycloak.Password.Value(); got != testSecrets["KEYCLOAK_PW"] {
		t.Errorf("keycloak password not usable: %q", got)
	}
	if got := cfg.Rancher2Config.AccessKey.Value(); got != testSecrets["RANCHER2_ACCESS_KEY"] {
		t.Errorf("access key not usable: %q", got)
	}
	if got := cfg.Rancher2Config.SecretKey.Value(); got != testSecrets["RANCHER2_SECRET_KEY"] {
		t.Errorf("secret key not usable: %q", got)
	}
}

func TestConfigLogMasksCredentialsFromEnv(t *testing.T) {
	for name, value := range testSecrets {
		t.Setenv(name, value)
	}
	cfg, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	assertMasked(t, cfg)
}

func TestConfigLogMasksCredentialsFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"keycloak":{"client_secret":"test-client-secret-1","password":"test-keycloak-password-2"},` +
		`"rancher2":{"access_key":"test-access-key-3","secret_key":"test-secret-key-4"}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	assertMasked(t, cfg)
}
