package registryprune

import (
	"sort"
	"strings"
	"time"
)

type Deploy struct {
	ID          string
	ComponentID string
	Tag         string
	Status      string
	CreatedAt   time.Time
	Live        bool
	InFlight    bool
}

type ComputeInput struct {
	Deploys      []Deploy
	RetainCounts map[string]int
	DefaultCount int
	LivePlans    []string
	GracePeriod  time.Duration
	Now          time.Time
}

const LegacyDeployTagPattern = `^dpl[0-9a-z]{23}$`

type TagSets struct {
	Owned []string
	Keep  []string
}

func Compute(in ComputeInput) TagSets {
	owned := map[string]bool{}
	keep := map[string]bool{}

	byComponent := map[string][]Deploy{}
	for _, d := range in.Deploys {
		protected := d.InFlight || d.Live || (in.GracePeriod > 0 && in.Now.Sub(d.CreatedAt) < in.GracePeriod)
		if protected && d.ID != "" {
			keep[d.ID] = true
		}
		if d.Tag == "" {
			continue
		}
		owned[d.Tag] = true
		byComponent[d.ComponentID] = append(byComponent[d.ComponentID], d)
		if protected {
			keep[d.Tag] = true
		}
	}

	for componentID, deploys := range byComponent {
		sort.SliceStable(deploys, func(i, j int) bool {
			return deploys[i].CreatedAt.After(deploys[j].CreatedAt)
		})
		k, ok := in.RetainCounts[componentID]
		if !ok {
			k = in.DefaultCount
		}
		seen := map[string]bool{}
		for _, d := range deploys {
			if len(seen) >= k {
				break
			}
			if seen[d.Tag] {
				continue
			}
			seen[d.Tag] = true
			keep[d.Tag] = true
		}
	}

	for tag := range owned {
		if keep[tag] {
			continue
		}
		for _, plan := range in.LivePlans {
			if strings.Contains(plan, tag) {
				keep[tag] = true
				break
			}
		}
	}

	return TagSets{Owned: sortedKeys(owned), Keep: sortedKeys(keep)}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
