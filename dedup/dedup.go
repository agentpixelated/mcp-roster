package dedup

import (
	"sort"

	"github.com/agentpixelated/mcp-roster/internal/compare"
	"github.com/agentpixelated/mcp-roster/model"
)

// FindDuplicates groups servers by name and reports duplicates.
func FindDuplicates(inv model.Inventory) []model.DuplicateGroup {
	nameMap := make(map[string][]model.MCPServer)
	for _, srv := range inv.Servers {
		nameMap[srv.Name] = append(nameMap[srv.Name], srv)
	}

	var groups []model.DuplicateGroup
	for name, servers := range nameMap {
		if len(servers) < 2 {
			continue
		}

		status := "identical"
		for i := 1; i < len(servers); i++ {
			if !compare.ServersIdentical(servers[0], servers[i]) {
				status = "divergent"
				break
			}
		}

		// Sort servers within group by client name for deterministic output
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].Source.Client < servers[j].Source.Client
		})

		groups = append(groups, model.DuplicateGroup{
			Name:    name,
			Servers: servers,
			Status:  status,
		})
	}

	// Sort groups by name for deterministic output
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	return groups
}