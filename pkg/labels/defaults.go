package labels

import "maps"

// ApplyDefaults reconciles an install's labels against its app's default
// labels. snapshot is the set of defaults previously applied — it is what lets
// a removed default be told apart from a user-set label. Templated defaults go
// to templates for later rendering; static defaults land in labels directly.
func ApplyDefaults(current, templates, snapshot, defaults Labels) (newLabels, newTemplates Labels, changed bool) {
	newLabels = make(Labels, len(current)+len(defaults))
	maps.Copy(newLabels, current)
	newTemplates = make(Labels, len(templates)+len(defaults))
	maps.Copy(newTemplates, templates)

	for key := range snapshot {
		if _, ok := defaults[key]; !ok {
			delete(newLabels, key)
			delete(newTemplates, key)
		}
	}

	for key, val := range defaults {
		if IsTemplatedValue(val) {
			newTemplates[key] = val
			continue
		}
		newLabels[key] = val
		delete(newTemplates, key)
	}

	changed = !maps.Equal(newLabels, current) || !maps.Equal(newTemplates, templates)
	return newLabels, newTemplates, changed
}
