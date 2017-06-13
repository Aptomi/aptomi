package slinga

import "reflect"

// LabelSet defines the set of labels that will be manipulated
type LabelSet struct {
	Labels map[string]string
}

// Apply set of transformations to labels
func (user *User) getLabelSet(dir string) LabelSet {
	secrets := LoadUserSecretsByIDFromDir(dir, user.ID)
	return LabelSet{Labels: user.Labels}.addLabels(LabelSet{Labels: secrets.Labels})
}

// Apply set of transformations to labels
func (src *LabelSet) applyTransform(ops *LabelOperations) LabelSet {
	result := LabelSet{Labels: make(map[string]string)}

	// copy original labels
	for k, v := range src.Labels {
		result.Labels[k] = v
	}

	if ops != nil {
		// set labels
		for k, v := range (*ops)["set"] {
			result.Labels[k] = v
		}

		// remove labels
		for k := range (*ops)["remove"] {
			delete(result.Labels, k)
		}
	}

	return result
}

// Merge two sets of labels
func (src LabelSet) addLabels(ops LabelSet) LabelSet {
	result := LabelSet{Labels: make(map[string]string)}

	// copy original labels
	for k, v := range src.Labels {
		result.Labels[k] = v
	}

	// put new labels
	for k, v := range ops.Labels {
		result.Labels[k] = v
	}

	return result
}

// Function to compare two labels sets. If one is nil and another one is empty, it will return true as well
func (src LabelSet) equal(dst LabelSet) bool {
	if len(src.Labels) == 0 && len(dst.Labels) == 0 {
		return true
	}
	return reflect.DeepEqual(src.Labels, dst.Labels)
}
