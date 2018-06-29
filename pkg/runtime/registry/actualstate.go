package registry

import (
	"fmt"

	"github.com/Aptomi/aptomi/pkg/engine/resolve"
)

func (reg *defaultRegistry) GetActualState() (*resolve.PolicyResolution, error) {
	actualState := resolve.NewPolicyResolution()

	//instances, err := reg.store.List(runtime.KeyFromParts(runtime.SystemNS, resolve.TypeComponentInstance.Kind, ""))
	var instances []*resolve.ComponentInstance
	err := reg.store.Find(resolve.TypeComponentInstance.Kind).List(&instances)
	if err != nil {
		return nil, fmt.Errorf("error while getting all component instances: %s", err)
	}

	for _, instance := range instances {
		key := instance.GetKey()
		actualState.ComponentInstanceMap[key] = instance
	}

	return actualState, nil
}