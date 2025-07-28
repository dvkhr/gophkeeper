package config

import (
	"fmt"
	"net"
)

// ValidateServerAddress проверяет, что адрес имеет формат host:port и порт валиден
func ValidateServerAddress(address string) error {
	if address == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid server address format '%s': must be host:port", address)
	}

	if host == "" {
		return fmt.Errorf("missing host in server address: %s", address)
	}

	portInt, err := net.LookupPort("tcp", port)
	if err != nil {
		return fmt.Errorf("invalid port in server address '%s': %v", address, err)
	}

	if portInt <= 0 || portInt > 65535 {
		return fmt.Errorf("port must be in range 1-65535, got %d", portInt)
	}

	return nil
}
