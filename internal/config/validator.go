package config

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

func (c *Config) Validate() error {
	validate := validator.New()

	if err := validate.RegisterValidation("port", validatePort); err != nil {
		return fmt.Errorf("port validation failed: %w", err)
	}
	if err := validate.RegisterValidation("host", validateHost); err != nil {
		return fmt.Errorf("host validation failed: %w", err)
	}
	_ = validate.RegisterValidation("hosts", validateHosts)

	if err := validate.RegisterValidation("address", validateAddress); err != nil {
		return fmt.Errorf("address validation failed: %w", err)
	}

	if err := validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func validatePort(fl validator.FieldLevel) bool {
	port := fl.Field().Int()
	return portValidation(int(port))
}

func validateHost(fl validator.FieldLevel) bool {
	host := fl.Field().String()
	return hostValidation(host)
}

func validateAddress(fl validator.FieldLevel) bool {
	rawAddresses := fl.Field().String()

	addresses := strings.Split(rawAddresses, ",")
	for _, address := range addresses {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return false
		}

		portInt, err := strconv.Atoi(port)
		if err != nil {
			return false
		}

		if !hostValidation(host) || !portValidation(portInt) {
			return false
		}
	}

	return true
}

func portValidation(port int) bool {
	return port >= 0 && port <= 65535
}

func hostValidation(host string) bool {
	// Check if ip address
	ip := net.ParseIP(host)
	if ip != nil {
		return true
	}

	// Check if hostname
	hostnameRegex := `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.?)+([A-Za-z]|[A-Za-z][A-Za-z0-9\-]*[A-Za-z0-9])$`
	matched, err := regexp.MatchString(hostnameRegex, host)
	if err != nil {
		return false
	}

	return matched
}

func validateHosts(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	hosts := strings.Split(value, ",")

	if len(hosts) == 0 {
		return false
	}

	for _, hostPort := range hosts {
		host, portStr, err := net.SplitHostPort(hostPort)
		if err != nil {
			return false
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return false
		}
		if portValidation(port) == false {
			return false
		}

		if hostValidation(host) == false {
			return false
		}
	}

	return true
}
