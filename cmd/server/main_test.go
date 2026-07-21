package main

import "testing"

func TestConfigFromEnvRequiresMySQLDSN(t *testing.T) {
	if _, err := configFromEnv(func(string) string { return "" }); err == nil {
		t.Fatal("configFromEnv() error = nil, want error")
	}
}

func TestConfigFromEnvUsesDefaultAddress(t *testing.T) {
	config, err := configFromEnv(func(key string) string {
		if key == "MYSQL_DSN" {
			return "tracker:password@tcp(mysql:3306)/tracker?parseTime=true"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("configFromEnv() error = %v", err)
	}
	if config.ListenAddress != ":8080" {
		t.Fatalf("ListenAddress = %q", config.ListenAddress)
	}
}
