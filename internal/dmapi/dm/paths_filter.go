package dm

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// CustomFilter is a named user-defined filter category containing specific DM type paths.
// When toggled hidden, its paths take priority over any parent base-type filter state.
type CustomFilter struct {
	Name   string
	Paths  []string
	Hidden bool
}

type PathsFilter struct {
	findDirectChildren func(string) []string
	filteredPaths      map[string]bool
	CustomFilters      []CustomFilter
}

func NewPathsFilter(findDirectChildren func(string) []string) *PathsFilter {
	return &PathsFilter{
		findDirectChildren: findDirectChildren,
		filteredPaths:      make(map[string]bool),
	}
}

func NewPathsFilterEmpty() *PathsFilter {
	return NewPathsFilter(func(string) []string {
		return nil
	})
}

func (p *PathsFilter) Clear() {
	p.filteredPaths = make(map[string]bool)
	for i := range p.CustomFilters {
		p.CustomFilters[i].Hidden = false
	}
}

func (p *PathsFilter) Copy() PathsFilter {
	filteredPaths := make(map[string]bool, len(p.filteredPaths))
	for path := range p.filteredPaths {
		filteredPaths[path] = true
	}
	customFilters := make([]CustomFilter, len(p.CustomFilters))
	copy(customFilters, p.CustomFilters)
	return PathsFilter{
		p.findDirectChildren,
		filteredPaths,
		customFilters,
	}
}

func (p *PathsFilter) IsHiddenPath(path string) bool {
	return p.filteredPaths[path]
}

func (p *PathsFilter) IsVisiblePath(path string) bool {
	return !p.IsHiddenPath(path)
}

func (p *PathsFilter) HasHiddenChildPath(path string) bool {
	for filteredPath := range p.filteredPaths {
		if strings.HasPrefix(filteredPath, path) {
			return true
		}
	}
	return false
}

func (p *PathsFilter) TogglePath(path string) {
	p.togglePath(path, p.IsVisiblePath(path))
	log.Printf("toggle [%s] path: [%t]", path, p.IsVisiblePath(path))
}

func (p *PathsFilter) togglePath(path string, isFilteredOut bool) {
	for _, directChild := range p.findDirectChildren(path) {
		p.togglePath(directChild, isFilteredOut)
	}
	if isFilteredOut {
		p.filteredPaths[path] = true
	} else {
		delete(p.filteredPaths, path)
	}
}

// AddCustomFilter adds a new custom filter category.
// The new filter starts visible (not hidden).
func (p *PathsFilter) AddCustomFilter(name string, paths []string) {
	p.CustomFilters = append(p.CustomFilters, CustomFilter{
		Name:   name,
		Paths:  paths,
		Hidden: false,
	})
	log.Printf("added custom filter [%s] with %d paths", name, len(paths))
}

// RemoveCustomFilter removes the custom filter at the given index.
// Its paths are cleared from filteredPaths so base filters resume control.
func (p *PathsFilter) RemoveCustomFilter(idx int) {
	if idx < 0 || idx >= len(p.CustomFilters) {
		return
	}
	f := p.CustomFilters[idx]
	for _, path := range f.Paths {
		p.clearFilteredSubtree(path)
	}
	p.CustomFilters = append(p.CustomFilters[:idx], p.CustomFilters[idx+1:]...)
	log.Printf("removed custom filter [%s]", f.Name)
}

// ToggleCustomFilter toggles the visibility of the custom filter at the given index.
func (p *PathsFilter) ToggleCustomFilter(idx int) {
	if idx < 0 || idx >= len(p.CustomFilters) {
		return
	}
	f := &p.CustomFilters[idx]
	f.Hidden = !f.Hidden
	for _, path := range f.Paths {
		p.togglePath(path, f.Hidden)
	}
	log.Printf("toggle custom filter [%s]: hidden=%t", f.Name, f.Hidden)
}

// clearFilteredSubtree removes a path and all its known children from filteredPaths.
func (p *PathsFilter) clearFilteredSubtree(path string) {
	delete(p.filteredPaths, path)
	for _, child := range p.findDirectChildren(path) {
		p.clearFilteredSubtree(child)
	}
}
