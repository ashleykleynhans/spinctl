package model

import (
	"fmt"
	"strings"
)

// MarshalYAML implements the yaml.Marshaler interface for ServiceName.
func (s ServiceName) MarshalYAML() (any, error) {
	if name, ok := serviceNameStrings[s]; ok {
		return name, nil
	}
	return nil, fmt.Errorf("unknown ServiceName: %d", int(s))
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ServiceName.
func (s *ServiceName) UnmarshalYAML(unmarshal func(any) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	name, err := ServiceNameFromString(str)
	if err != nil {
		return err
	}
	*s = name
	return nil
}

type ServiceName int

const (
	Clouddriver ServiceName = iota + 1
	Orca
	Gate
	Front50
	Echo
	Igor
	Fiat
	Rosco
	Kayenta
	Deck
)

var serviceNameStrings = map[ServiceName]string{
	Clouddriver: "clouddriver",
	Orca:        "orca",
	Gate:        "gate",
	Front50:     "front50",
	Echo:        "echo",
	Igor:        "igor",
	Fiat:        "fiat",
	Rosco:       "rosco",
	Kayenta:     "kayenta",
	Deck:        "deck",
}

var stringToServiceName map[string]ServiceName

func init() {
	stringToServiceName = make(map[string]ServiceName, len(serviceNameStrings))
	for k, v := range serviceNameStrings {
		stringToServiceName[v] = k
	}
}

func (s ServiceName) String() string {
	if name, ok := serviceNameStrings[s]; ok {
		return name
	}
	return fmt.Sprintf("ServiceName(%d)", int(s))
}

func ServiceNameFromString(s string) (ServiceName, error) {
	lower := strings.ToLower(strings.TrimSpace(s))
	if name, ok := stringToServiceName[lower]; ok {
		return name, nil
	}
	return 0, fmt.Errorf("unknown service name: %q", s)
}

func AllServiceNames() []ServiceName {
	return []ServiceName{
		Clouddriver, Orca, Gate, Front50, Echo,
		Igor, Fiat, Rosco, Kayenta, Deck,
	}
}

func (s ServiceName) PackageName() string {
	return "spinnaker-" + s.String()
}

func (s ServiceName) SystemdUnit() string {
	return s.String() + ".service"
}

func (s ServiceName) ConfigFile() string {
	return s.String() + ".yml"
}

func DeploymentOrder() [][]ServiceName {
	return [][]ServiceName{
		{Front50},
		{Fiat},
		{Clouddriver},
		{Orca},
		{Echo},
		{Igor, Rosco, Kayenta},
		{Gate},
		{Deck},
	}
}
