package multiminer

import "testing"

func TestAddressValidator(t *testing.T) {
	validator := NewAddressValidator()
	
	tests := []struct {
		address string
		valid   bool
		desc    string
	}{
		{"192.168.1.100:4028", true, "valid private IP with allowed port"},
		{"localhost:4028", true, "localhost with allowed port"},
		{"example.com:8080", true, "hostname with allowed port"},
		{"192.168.1.100:1234", false, "private IP with disallowed port"},
		{"", false, "empty address"},
		{"192.168.1.100", false, "missing port"},
		{"192.168.1.100:invalid", false, "invalid port"},
		{"http://192.168.1.100", true, "HTTP URL"},
		{"https://example.com/api", true, "HTTPS URL"},
	}
	
	for _, test := range tests {
		err := validator.ValidateAddress(test.address)
		isValid := err == nil
		
		if isValid != test.valid {
			t.Errorf("Address %q: expected valid=%v, got valid=%v (%s)", 
				test.address, test.valid, isValid, test.desc)
			if err != nil {
				t.Logf("Error: %v", err)
			}
		}
	}
}

func TestCommandValidator(t *testing.T) {
	validator := NewCommandValidator()
	
	validCommands := []string{"version", "summary", "devs", "pools", "stats"}
	invalidCommands := []string{"", "rm", "shutdown", "delete", "format"}
	
	for _, cmd := range validCommands {
		err := validator.ValidateCommand(cmd)
		if err != nil {
			t.Errorf("Command %q should be valid, got error: %v", cmd, err)
		}
	}
	
	for _, cmd := range invalidCommands {
		err := validator.ValidateCommand(cmd)
		if err == nil {
			t.Errorf("Command %q should be invalid", cmd)
		}
	}
}

func TestParameterValidator(t *testing.T) {
	validator := NewCommandValidator()
	
	validParams := []string{"", "1", "pool.example.com:4242", "user123"}
	invalidParams := []string{
		"param;injection",
		"param&injection", 
		"param|injection",
		"param`injection",
		"param$injection",
		"param(injection)",
		"param<injection>",
	}
	
	for _, param := range validParams {
		err := validator.ValidateParameter("version", param)
		if err != nil {
			t.Errorf("Parameter %q should be valid, got error: %v", param, err)
		}
	}
	
	for _, param := range invalidParams {
		err := validator.ValidateParameter("version", param)
		if err == nil {
			t.Errorf("Parameter %q should be invalid", param)
		}
	}
	
	// Test parameter length limit
	longParam := string(make([]byte, 1001))
	err := validator.ValidateParameter("version", longParam)
	if err == nil {
		t.Error("Long parameter should be invalid")
	}
}