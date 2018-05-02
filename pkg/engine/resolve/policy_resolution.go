package resolve

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// PolicyResolution contains resolution data for the policy. It essentially represents the desired state calculated
// by policy resolver. It contains a calculated map of component instances with their data, information about
// resolved service consumption declarations, as well as processing order to components in which they have to be
// instantiated/updated/deleted.
type PolicyResolution struct {
	// Desired state is generated by resolver and has all fields populated.
	// Actual state is usually loaded from the store and only contains component instance map.
	isDesired bool

	// Resolved component instances: componentKey -> componentInstance
	ComponentInstanceMap map[string]*ComponentInstance

	// Resolved dependencies: dependencyID -> serviceKey
	dependencyInstanceMap map[string]*DependencyResolution

	// Resolved component processing order in which components/services have to be processed
	componentProcessingOrderHas map[string]bool
	componentProcessingOrder    []string
}

// NewPolicyResolution creates new empty PolicyResolution, given a flag indicating whether it's a
// desired state (generated by a resolver), or actual state (loaded from the store)
func NewPolicyResolution(isDesired bool) *PolicyResolution {
	return &PolicyResolution{
		isDesired:                   isDesired,
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		dependencyInstanceMap:       make(map[string]*DependencyResolution),
		componentProcessingOrderHas: make(map[string]bool),
		componentProcessingOrder:    []string{},
	}
}

// GetComponentInstanceEntry retrieves a component instance entry by key, or creates an new entry if it doesn't exist
func (resolution *PolicyResolution) GetComponentInstanceEntry(cik *ComponentInstanceKey) *ComponentInstance {
	key := cik.GetKey()
	if _, ok := resolution.ComponentInstanceMap[key]; !ok {
		resolution.ComponentInstanceMap[key] = newComponentInstance(cik)
	}
	return resolution.ComponentInstanceMap[key]
}

// RecordResolved takes a component instance and adds a new dependency record into it
func (resolution *PolicyResolution) RecordResolved(cik *ComponentInstanceKey, dependency *lang.Dependency, ruleResult *lang.RuleActionResult) {
	instance := resolution.GetComponentInstanceEntry(cik)
	instance.addDependency(runtime.KeyForStorable(dependency))
	instance.addRuleInformation(ruleResult)
	resolution.recordProcessingOrder(cik)
}

// Record processing order for component instance
func (resolution *PolicyResolution) recordProcessingOrder(cik *ComponentInstanceKey) {
	key := cik.GetKey()
	if !resolution.componentProcessingOrderHas[key] {
		resolution.componentProcessingOrderHas[key] = true
		resolution.componentProcessingOrder = append(resolution.componentProcessingOrder, key)
	}
}

// RecordCodeParams stores calculated code params for component instance
func (resolution *PolicyResolution) RecordCodeParams(cik *ComponentInstanceKey, codeParams util.NestedParameterMap) error {
	return resolution.GetComponentInstanceEntry(cik).addCodeParams(codeParams)
}

// RecordDiscoveryParams stores calculated discovery params for component instance
func (resolution *PolicyResolution) RecordDiscoveryParams(cik *ComponentInstanceKey, discoveryParams util.NestedParameterMap) error {
	return resolution.GetComponentInstanceEntry(cik).addDiscoveryParams(discoveryParams)
}

// RecordLabels stores calculated labels for component instance
func (resolution *PolicyResolution) RecordLabels(cik *ComponentInstanceKey, labels *lang.LabelSet) {
	resolution.GetComponentInstanceEntry(cik).addLabels(labels)
}

// StoreEdge stores incoming/outgoing graph edges for component instance for observability and reporting
func (resolution *PolicyResolution) StoreEdge(src *ComponentInstanceKey, dst *ComponentInstanceKey) {
	// Arrival key can be empty at the very top of the recursive function in engine, so let's check for that
	if src != nil && dst != nil {
		resolution.GetComponentInstanceEntry(src).addEdgeOut(dst.GetKey())
		resolution.GetComponentInstanceEntry(dst).addEdgeIn(src.GetKey())
	}
}

// AppendData appends data to the current PolicyResolution record by aggregating data over component instances.
// If there is a conflict (e.g. components have different code parameters), then an error will be reported.
func (resolution *PolicyResolution) AppendData(ops *PolicyResolution) error {
	for _, instance := range ops.ComponentInstanceMap {
		err := resolution.GetComponentInstanceEntry(instance.Metadata.Key).appendData(instance)
		if err != nil {
			return err
		}
	}
	for key := range ops.ComponentInstanceMap {
		resolution.recordProcessingOrder(ops.ComponentInstanceMap[key].Metadata.Key)
	}
	return nil
}

// GetComponentProcessingOrder returns component processing order, as determined by the policy resolver, in which
// components/services have to be processed
func (resolution *PolicyResolution) GetComponentProcessingOrder() []string {
	if !resolution.isDesired {
		panic("attempting to get component processing order from actual state")
	}
	return resolution.componentProcessingOrder
}

// GetDependencyInstanceMap returns map which contains resolution status for every dependency
func (resolution *PolicyResolution) GetDependencyInstanceMap() map[string]*DependencyResolution {
	if !resolution.isDesired {
		panic("attempting to get dependency instance map for actual state")
	}
	return resolution.dependencyInstanceMap
}

// SetDependencyInstanceMap overrides existing dependencyInstanceMap
func (resolution *PolicyResolution) SetDependencyInstanceMap(dMap map[string]*DependencyResolution) {
	// TODO: we actually need to start saving dependencyInstanceMap into the store. after that we can delete this method
	resolution.dependencyInstanceMap = dMap
}

// Validate checks that the state is valid, meaning that all objects references are valid. It takes all the instances
// and verifies that all services exist, all clusters exist, etc
func (resolution *PolicyResolution) Validate(policy *lang.Policy) error {
	// component instances must point to valid objects
	for _, instance := range resolution.ComponentInstanceMap {
		componentKey := instance.Metadata.Key

		// verify that contract exists
		contractObj, err := policy.GetObject(lang.ContractObject.Kind, componentKey.ContractName, componentKey.Namespace)
		if contractObj == nil || err != nil {
			// component instance points to non-existing contract, meaning this component instance is now orphan
			return fmt.Errorf("contract '%s/%s' can only be deleted after it's no longer in use. still used by: %s", componentKey.Namespace, componentKey.ContractName, componentKey.GetKey())
		}

		// verify that context within a contract exists
		contract := contractObj.(*lang.Contract)
		contextExists := false
		for _, context := range contract.Contexts {
			if context.Name == componentKey.ContextName {
				contextExists = true
				break
			}
		}
		if !contextExists {
			// component instance points to non-existing context within a contract, meaning this component instance is now orphan
			return fmt.Errorf("context '%s/%s/%s' can only be deleted after it's no longer in use. still used by: %s", componentKey.Namespace, componentKey.ContractName, componentKey.ContextName, componentKey.GetKey())
		}

		// verify that service exists
		serviceObj, err := policy.GetObject(lang.ServiceObject.Kind, componentKey.ServiceName, componentKey.Namespace)
		if serviceObj == nil || err != nil {
			// component instance points to non-existing service, meaning this component instance is now orphan
			return fmt.Errorf("service '%s/%s' can only be deleted after it's no longer in use. still used by: %s", componentKey.Namespace, componentKey.ServiceName, componentKey.GetKey())
		}

		if componentKey.ComponentName != componentRootName {
			// verify that component within a service exists
			service := serviceObj.(*lang.Service)
			component, found := service.GetComponentsMap()[componentKey.ComponentName]
			if component == nil || !found {
				// component instance points to non-existing component within a service, meaning this component instance is now orphan
				return fmt.Errorf("component '%s/%s/%s' can only be deleted after it's no longer in use. still used by: %s", componentKey.Namespace, componentKey.ServiceName, componentKey.ComponentName, componentKey.GetKey())
			}
		}

		// verify that cluster exists
		clusterObj, err := policy.GetObject(lang.ClusterObject.Kind, componentKey.ClusterName, runtime.SystemNS)
		if clusterObj == nil || err != nil {
			// component instance points to non-existing cluster, meaning this component instance is now orphan
			return fmt.Errorf("cluster '%s/%s' can only be deleted after it's no longer in use. still used by: %s", componentKey.Namespace, componentKey.ClusterName, componentKey.GetKey())
		}
	}
	return nil
}

// AllDependenciesResolvedSuccessfully returns if all dependencies got resolved successfully
func (resolution *PolicyResolution) AllDependenciesResolvedSuccessfully() bool {
	return resolution.SuccessfullyResolvedDependencies() == len(resolution.dependencyInstanceMap)
}

// SuccessfullyResolvedDependencies returns the number of successfully resolved dependencies
func (resolution *PolicyResolution) SuccessfullyResolvedDependencies() int {
	result := 0
	for _, d := range resolution.dependencyInstanceMap {
		if d.Resolved {
			result++
		}
	}
	return result
}
