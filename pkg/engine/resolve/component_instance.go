package resolve

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Aptomi/aptomi/pkg/errors"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/util"
)

// ComponentInstanceObject is an informational data structure with Kind and Constructor for component instance object
var ComponentInstanceObject = &runtime.Info{
	Kind:        "component-instance",
	Storable:    true,
	Versioned:   false,
	Constructor: func() runtime.Object { return &ComponentInstance{} },
}

// ComponentInstanceMetadata is object metadata for ComponentInstance
type ComponentInstanceMetadata struct {
	Key *ComponentInstanceKey
}

// AllowIngres is an special key, which is used in DataForPlugins to indicate whether ingress traffic should be allowed for a given component instance
const AllowIngres = "allow_ingress"

// ComponentInstance is an instance of a particular code component within a service, which indicate that this component
// has to be instantiated and configured in a certain cluster. Policy resolver produces a map of component instances
// and their parameters in desired state (PolicyResolution) as result of policy resolution.
//
// When a service gets instantiated, a special "root" component instance gets created with component name
// set to Metadata.Key.ComponentName == componentRootName. Then all "child" component instances get created, one
// per service component with type code.
//
// Every ComponentInstance contains full information about for a given component instance - who is consuming this
// instance, labels, code params, discovery params, in & out graph edges (what component instances we depend on,
// what component instances depend on us), creation/update times, and endpoints retrieved from the underlying cloud.
type ComponentInstance struct {
	/*
		These fields get populated during policy resolution as a part of desired state.
		When adding new fields to this object, it's crucial to modify appendData() method as well (!).
	*/

	runtime.TypeKind `yaml:",inline"`

	// Metadata is an object metadata for component instance
	Metadata *ComponentInstanceMetadata

	// Error means that a critical error happened with a component (e.g. conflict of parameters)
	Error error

	// DependencyKeys is a list of dependency keys which are keeping this component instantiated (if dependency resolves to this component directly, then the value is 0. otherwise it's depth in policy resolution)
	DependencyKeys map[string]int

	// IsCode means the component is code
	IsCode bool

	// CalculatedLabels is a set of calculated labels for the component, aggregated over all uses
	CalculatedLabels *lang.LabelSet

	// CalculatedDiscovery is a set of calculated discovery parameters for the component (non-conflicting over all uses of this component)
	CalculatedDiscovery util.NestedParameterMap

	// CalculatedCodeParams is a set of calculated code parameters for the component (non-conflicting over all uses of this component)
	CalculatedCodeParams util.NestedParameterMap

	// DataForPlugins is an additional data recorded for use in plugins
	DataForPlugins map[string]string

	/*
		These fields only make sense for the desired state. They will NOT be present in actual state
	*/

	// EdgesOut is a set of outgoing graph edges ('key' -> true) from this component instance. Only makes sense as a part of desired state
	EdgesOut map[string]bool

	/*
		These fields only make sense for the actual state. They will NOT be present in desired state
	*/

	// CreatedAt is when this component instance was created
	CreatedAt time.Time

	// UpdatedAt is the last time when this component instance was updated
	UpdatedAt time.Time

	// EndpointsUpToDate is true if endpoints are up to date and false otherwise
	EndpointsUpToDate bool

	// Endpoints represents all URLs that could be used to access deployed service
	Endpoints map[string]string
}

// Creates a new component instance
func newComponentInstance(cik *ComponentInstanceKey) *ComponentInstance {
	return &ComponentInstance{
		TypeKind:             ComponentInstanceObject.GetTypeKind(),
		Metadata:             &ComponentInstanceMetadata{Key: cik},
		DependencyKeys:       make(map[string]int),
		CalculatedLabels:     lang.NewLabelSet(make(map[string]string)),
		CalculatedDiscovery:  util.NestedParameterMap{},
		CalculatedCodeParams: util.NestedParameterMap{},
		EdgesOut:             make(map[string]bool),
		DataForPlugins:       make(map[string]string),
		Endpoints:            make(map[string]string),
	}
}

// GetKey returns an object key. It's a component instance key, as generated by the engine
func (instance *ComponentInstance) GetKey() string {
	return instance.Metadata.Key.GetKey()
}

// GetDeployName returns a string that could be used as name for deployment inside the cluster
func (instance *ComponentInstance) GetDeployName() string {
	return instance.Metadata.Key.GetDeployName()
}

// GetNamespace returns an object namespace. It's a system namespace for all component instances
func (instance *ComponentInstance) GetNamespace() string {
	return runtime.SystemNS
}

// GetName returns object name
func (instance *ComponentInstance) GetName() string {
	return instance.GetKey()
}

// GetRunningTime returns the total lifetime of a component instance (since it was launched till now)
func (instance *ComponentInstance) GetRunningTime() time.Duration {
	return time.Since(instance.CreatedAt)
}

func (instance *ComponentInstance) addDependency(dependencyKey string, depth int) {
	instance.DependencyKeys[dependencyKey] = depth
}

func (instance *ComponentInstance) addRuleInformation(result *lang.RuleActionResult) {
	instance.DataForPlugins[AllowIngres] = strconv.FormatBool(!result.RejectIngress)
}

func (instance *ComponentInstance) addCodeParams(codeParams util.NestedParameterMap) error {
	if len(instance.CalculatedCodeParams) == 0 {
		// Record code parameters
		instance.CalculatedCodeParams = codeParams
	} else if !instance.CalculatedCodeParams.DeepEqual(codeParams) {
		// Same component instance, different code parameters
		return errors.NewErrorWithDetails(
			fmt.Sprintf("conflicting code parameters for component instance: %s", instance.GetKey()),
			errors.Details{
				"code_params_existing": instance.CalculatedCodeParams,
				"code_params_new":      codeParams,
				"diff":                 instance.CalculatedCodeParams.Diff(codeParams),
			},
		)
	}
	return nil
}

func (instance *ComponentInstance) addDiscoveryParams(discoveryParams util.NestedParameterMap) error {
	if len(instance.CalculatedDiscovery) == 0 {
		// Record discovery parameters
		instance.CalculatedDiscovery = discoveryParams
	} else if !instance.CalculatedDiscovery.DeepEqual(discoveryParams) {
		// Same component instance, different discovery parameters
		return errors.NewErrorWithDetails(
			fmt.Sprintf("conflicting discovery parameters for component instance: %s", instance.GetKey()),
			errors.Details{
				"discovery_params_existing": instance.CalculatedDiscovery,
				"discovery_params_new":      discoveryParams,
				"diff":                      instance.CalculatedDiscovery.Diff(discoveryParams),
			},
		)
	}
	return nil
}

func (instance *ComponentInstance) addLabels(labels *lang.LabelSet) {
	// it's pretty typical for us to come with different labels to a component instance, let's combine them all
	instance.CalculatedLabels.AddLabels(labels.Labels)
}

func (instance *ComponentInstance) addEdgeOut(dstKey string) {
	instance.EdgesOut[dstKey] = true
}

// UpdateTimes updates component creation and update times
func (instance *ComponentInstance) UpdateTimes(createdAt time.Time, updatedAt time.Time) {
	if time.Time.IsZero(instance.CreatedAt) || (!time.Time.IsZero(createdAt) && createdAt.Before(instance.CreatedAt)) {
		instance.CreatedAt = createdAt
	}
	if !time.Time.IsZero(updatedAt) && updatedAt.After(instance.UpdatedAt) {
		instance.UpdatedAt = updatedAt
	}
}

// appendData gets called to append data for two existing component instances, both of which have been already processed
// and populated with data
func (instance *ComponentInstance) appendData(ops *ComponentInstance) {
	// Combine dependencies which are keeping this component instantiated
	for dependencyKey, depth := range ops.DependencyKeys {
		instance.addDependency(dependencyKey, depth)
	}

	// Transfer IsCode bool
	if instance.IsCode != ops.IsCode {
		instance.Error = fmt.Errorf("component %s can't be converted from code to non-code and vice versa", instance.GetKey())
		return
	}
	instance.IsCode = instance.IsCode || ops.IsCode

	// Combine labels
	instance.addLabels(ops.CalculatedLabels)

	// Combine discovery params
	var err = instance.addDiscoveryParams(ops.CalculatedDiscovery)
	if err != nil {
		instance.Error = err
		return
	}

	// Combine code params
	err = instance.addCodeParams(ops.CalculatedCodeParams)
	if err != nil {
		instance.Error = err
		return
	}

	// Outgoing graph edges (instance: key -> true) as we are traversing the graph
	for keyDst := range ops.EdgesOut {
		instance.addEdgeOut(keyDst)
	}

	// Data for plugins
	for k, v := range ops.DataForPlugins {
		instance.DataForPlugins[k] = v
	}
}
