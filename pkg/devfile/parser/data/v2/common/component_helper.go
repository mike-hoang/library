package common

import (
	"fmt"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

// IsContainer checks if the component is a container
func IsContainer(component v1.Component) bool {
	if component.Container != nil {
		return true
	}
	return false
}

// IsVolume checks if the component is a volume
func IsVolume(component v1.Component) bool {
	if component.Volume != nil {
		return true
	}
	return false
}

// GetComponentType returns the component type of a given component
func GetComponentType(component v1.Component) (v1.ComponentType, error) {
	switch {
	case component.Container != nil:
		return v1.ContainerComponentType, nil
	case component.Volume != nil:
		return v1.VolumeComponentType, nil
	case component.Plugin != nil:
		return v1.PluginComponentType, nil
	case component.Kubernetes != nil:
		return v1.KubernetesComponentType, nil
	case component.Openshift != nil:
		return v1.OpenshiftComponentType, nil
	case component.Custom != nil:
		return v1.CustomComponentType, nil

	default:
		return "", fmt.Errorf("unknown component type")
	}
}
