package lang

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang/expression"
	"github.com/Aptomi/aptomi/pkg/object"
	"sync"
)

// ACLResolver is a struct which allows to perform ACL resolution, allowing to retrieve user privileges for the
// objects they access
type ACLResolver struct {
	rules        []*ACLRule
	cache        *expression.Cache
	roleMapCache sync.Map
}

// NewACLResolver creates a new ACLResolver
func NewACLResolver(globalRules *GlobalRules) *ACLResolver {
	return &ACLResolver{
		rules:        globalRules.GetRulesSortedByWeight(),
		cache:        expression.NewCache(),
		roleMapCache: sync.Map{},
	}
}

// GetUserPrivileges is a main method which determines privileges that a given user has for a given object
func (resolver *ACLResolver) GetUserPrivileges(user *User, obj object.Base) (*Privilege, error) {
	roleMap, err := resolver.getUserRoleMap(user)
	if err != nil {
		return nil, err
	}

	// figure out which role's privileges apply
	for _, role := range ACLRolesOrderedList {
		namespaceSpan := roleMap[role.ID]
		if namespaceSpan[namespaceAll] || namespaceSpan[obj.GetNamespace()] {
			return role.Privileges.getObjectPrivileges(obj), nil
		}
	}

	return nobody.Privileges.getObjectPrivileges(obj), nil
}

// Returns privileges for a given object
func (privileges *Privileges) getObjectPrivileges(obj object.Base) *Privilege {
	var result *Privilege
	if obj.GetNamespace() == object.SystemNS {
		result = privileges.GlobalObjects[obj.GetKind()]
	} else {
		result = privileges.NamespaceObjects[obj.GetKind()]
	}
	if result == nil {
		return noAccess
	}
	return result
}

// Returns the map role ID -> to which namespaces this role applies
// Note that user may have multiple roles at the same time. E.g.
// - domain admin (i.e. for all namespaces within Aptomi domain)
// - namespace admin for a set of given namespaces
// - service consumer for a set of given namespaces
func (resolver *ACLResolver) getUserRoleMap(user *User) (map[string]map[string]bool, error) {
	roleMapCached, ok := resolver.roleMapCache.Load(user.ID)
	if ok {
		return roleMapCached.(map[string]map[string]bool), nil
	}

	result := NewRuleActionResult(NewLabelSet(make(map[string]string)))
	if user.Admin {
		// this user is explicitly specified as domain admin
		result.RoleMap[domainAdmin.ID] = make(map[string]bool)
		result.RoleMap[domainAdmin.ID][namespaceAll] = true
	} else {
		// we need to run this user through ACL list
		params := expression.NewParams(user.Labels, nil)
		for _, rule := range resolver.rules {
			matched, err := rule.Matches(params, resolver.cache)
			if err != nil {
				return nil, fmt.Errorf("unable to resolve role for user '%s': %s", user.ID, err)
			}
			if matched {
				rule.ApplyActions(result)
			}
		}
	}

	resolver.roleMapCache.Store(user.ID, result.RoleMap)
	return result.RoleMap, nil
}