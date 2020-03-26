package envoy

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-sockaddr/template"
)

const defaultMeshGatewayPort int = 443

// ServiceAddressValue implements a flag.Value that may be used to parse an
// addr:port string into an api.ServiceAddress.
type ServiceAddressValue struct {
	value api.ServiceAddress
}

func (s *ServiceAddressValue) String() string {
	if s == nil {
		return fmt.Sprintf(":%d", defaultMeshGatewayPort)
	}
	return fmt.Sprintf("%v:%d", s.value.Address, s.value.Port)
}

func (s *ServiceAddressValue) Value() api.ServiceAddress {
	if s == nil || s.value.Port == 0 && s.value.Address == "" {
		return api.ServiceAddress{Port: defaultMeshGatewayPort}
	}
	return s.value
}

func (s *ServiceAddressValue) Set(raw string) error {
	var err error
	s.value, err = parseAddress(raw)
	return err
}

func parseAddress(raw string) (api.ServiceAddress, error) {
	result := api.ServiceAddress{}
	x, err := template.Parse(raw)
	if err != nil {
		return result, fmt.Errorf("Error parsing address %q: %v", raw, err)
	}

	addr, portStr, err := net.SplitHostPort(x)
	if err != nil {
		return result, fmt.Errorf("Error parsing address %q: %v", x, err)
	}

	port := defaultMeshGatewayPort
	if portStr != "" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return result, fmt.Errorf("Error parsing port %q: %v", portStr, err)
		}
	}

	result.Address = addr
	result.Port = port
	return result, nil
}

var _ flag.Value = (*ServiceAddressValue)(nil)

type ServiceAddressMapValue struct {
	value map[string]api.ServiceAddress
}

func (s *ServiceAddressMapValue) String() string {
	buf := new(strings.Builder)
	for k, v := range s.value {
		buf.WriteString(fmt.Sprintf("%v=%v:%d,", k, v.Address, v.Port))
	}
	return buf.String()
}

func (s *ServiceAddressMapValue) Set(raw string) error {
	if s.value == nil {
		s.value = make(map[string]api.ServiceAddress)
	}
	idx := strings.Index(raw, "=")
	if idx == -1 {
		return fmt.Errorf(`Missing "=" in argument: %s`, raw)
	}
	key, value := raw[0:idx], raw[idx+1:]
	var err error
	s.value[key], err = parseAddress(value)
	return err
}

var _ flag.Value = (*ServiceAddressMapValue)(nil)