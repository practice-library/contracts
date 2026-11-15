package atinventory

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
)

// closure adds to out the pieces reachable from p through next (p too, if it is on a cycle).
func closure(p string, next map[string][]string, out map[string]bool) map[string]bool {
	for _, n := range next[p] {
		if !out[n] {
			out[n] = true
			closure(n, next, out)
		}
	}
	return out
}

// Affected returns the sorted pieces touched by the changed paths, plus every piece that
// transitively consumes one (consumes, testConsumes, generatorConsumes). It runs no git. A path
// in a piece's directory affects that piece; any other path (root files, .github/, policy/,
// inventory.json, an unlisted pieces/ directory) affects every piece. Only path-form entries are
// edges: a schema-id entry names no piece in W0 and is ignored. A cycle, or an edge to a path
// not in the inventory, is an error.
func Affected(inv map[string]any, changed []string) ([]string, error) {
	list, ok := inv["pieces"].([]any)
	if !ok {
		return nil, fmt.Errorf("inventory: no pieces list")
	}
	deps := map[string][]string{} // piece -> pieces it consumes
	for _, e := range list {
		m, _ := e.(map[string]any)
		p, _ := m["path"].(string)
		piece, _ := m["piece"].(map[string]any)
		if p == "" || piece == nil {
			return nil, fmt.Errorf("inventory: malformed entry")
		}
		deps[p] = nil
		for _, k := range []string{"consumes", "testConsumes", "generatorConsumes"} {
			items, _ := piece[k].([]any)
			for _, it := range items {
				if s, _ := it.(string); strings.HasPrefix(s, "pieces/") {
					deps[p] = append(deps[p], s)
				}
			}
		}
	}
	var all []string
	users := map[string][]string{} // piece -> pieces that consume it
	for p, ds := range deps {
		all = append(all, p)
		for _, d := range ds {
			if _, ok := deps[d]; !ok {
				return nil, fmt.Errorf("%s consumes %s, which is not in the inventory", p, d)
			}
			users[d] = append(users[d], p)
		}
	}
	sort.Strings(all)
	for _, p := range all {
		if closure(p, deps, map[string]bool{})[p] {
			return nil, fmt.Errorf("consumes cycle through %s", p)
		}
	}
	seen := map[string]bool{}
	for _, c := range changed {
		c = path.Clean(c)
		i := slices.IndexFunc(all, func(p string) bool { return c == p || strings.HasPrefix(c, p+"/") })
		if i < 0 {
			return all, nil
		}
		seen[all[i]] = true
		closure(all[i], users, seen)
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}
