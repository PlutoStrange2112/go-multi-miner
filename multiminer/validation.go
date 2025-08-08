package multiminer

import (
	"net"
	"regexp"
	"strconv"
	"strings"
)

// AddressValidator validates device addresses
type AddressValidator struct {
	allowedPorts []int
	allowedHosts []string
}

// NewAddressValidator creates a new address validator
func NewAddressValidator() *AddressValidator {
	return &AddressValidator{
		allowedPorts: []int{4028, 8080, 80, 443, 8000, 8080, 9090, 3000, 4029},
	}
}

// ValidateAddress validates that an address is safe to connect to
func (v *AddressValidator) ValidateAddress(address string) error {
	if address == "" {
		return NewInvalidInputError("address cannot be empty")
	}

	// Handle HTTP URLs
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		return v.validateHTTPAddress(address)
	}

	// Handle host:port format
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return NewInvalidInputError("invalid address format, expected host:port")
	}

	// Validate host
	if err := v.validateHost(host); err != nil {
		return err
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return NewInvalidInputError("invalid port number")
	}

	if !v.isPortAllowed(port) {
		return NewInvalidInputError("port not in allowed list")
	}

	return nil
}

// validateHost checks if a host is valid and safe
func (v *AddressValidator) validateHost(host string) error {
	if host == "" {
		return NewInvalidInputError("host cannot be empty")
	}

	// Check for localhost variations (could be security risk in some environments)
	if isLocalhost(host) {
		return nil // Allow localhost for development
	}

	// Validate IP addresses
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return nil // Allow private IPs
		}
		// For public IPs, you might want additional validation
		return nil
	}

	// Validate hostnames
	if !isValidHostname(host) {
		return NewInvalidInputError("invalid hostname format")
	}

	return nil
}

// validateHTTPAddress validates HTTP/HTTPS URLs
func (v *AddressValidator) validateHTTPAddress(address string) error {
	// Basic URL validation - in production, you'd want more sophisticated validation
	if !strings.Contains(address, "://") {
		return NewInvalidInputError("invalid HTTP URL format")
	}

	// Extract host from URL for additional validation
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(address, "http://"), "https://"), "/")
	if len(parts) == 0 {
		return NewInvalidInputError("cannot parse host from URL")
	}

	host := parts[0]
	if strings.Contains(host, ":") {
		hostPart, _, err := net.SplitHostPort(host)
		if err != nil {
			return NewInvalidInputError("invalid URL host:port format")
		}
		host = hostPart
	}

	return v.validateHost(host)
}

// isPortAllowed checks if a port is in the allowed list
func (v *AddressValidator) isPortAllowed(port int) bool {
	for _, allowedPort := range v.allowedPorts {
		if port == allowedPort {
			return true
		}
	}
	return false
}

// isLocalhost checks if the host is a localhost variant
func isLocalhost(host string) bool {
	localhosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}
	for _, lh := range localhosts {
		if host == lh {
			return true
		}
	}
	return false
}

// isPrivateIP checks if an IP is in private ranges
func isPrivateIP(ip net.IP) bool {
	privateIPBlocks := []*net.IPNet{
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// isValidHostname validates hostname format
func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// Simple hostname validation - allows alphanumeric and dashes
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	return hostnameRegex.MatchString(hostname)
}

// CommandValidator validates miner commands for security
type CommandValidator struct {
	allowedCommands map[string]bool
}

// NewCommandValidator creates a new command validator
func NewCommandValidator() *CommandValidator {
	allowed := map[string]bool{
		"version":     true,
		"summary":     true,
		"devs":        true,
		"pools":       true,
		"stats":       true,
		"addpool":     true,
		"enablepool":  true,
		"disablepool": true,
		"removepool":  true,
		"switchpool":  true,
		"restart":     true,
		"quit":        true,
		"config":      true,
		"lcd":         true,
		"fans":        true,
		"temps":       true,
	}

	return &CommandValidator{allowedCommands: allowed}
}

// ValidateCommand checks if a command is allowed
func (v *CommandValidator) ValidateCommand(command string) error {
	if command == "" {
		return NewInvalidInputError("command cannot be empty")
	}

	if !v.allowedCommands[strings.ToLower(command)] {
		return NewInvalidInputError("command not allowed")
	}

	return nil
}

// ValidateParameter validates command parameters
func (v *CommandValidator) ValidateParameter(command, parameter string) error {
	// Basic parameter validation - extend as needed per command
	if len(parameter) > 1000 {
		return NewInvalidInputError("parameter too long")
	}

	// Prevent command injection
	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerousChars {
		if strings.Contains(parameter, char) {
			return NewInvalidInputError("parameter contains dangerous characters")
		}
	}

	return nil
}