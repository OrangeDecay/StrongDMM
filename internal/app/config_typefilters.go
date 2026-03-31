package app

import "github.com/rs/zerolog/log"

const (
	typeFiltersConfigName    = "typefilters"
	typeFiltersConfigVersion = 1
)

// savedCustomFilter stores the persistent definition of a custom type filter.
// The Hidden state is intentionally not saved — filters start visible on each launch.
type savedCustomFilter struct {
	Name  string
	Paths []string
}

type typeFiltersConfig struct {
	Version uint

	CustomFilters []savedCustomFilter
}

func (typeFiltersConfig) Name() string {
	return typeFiltersConfigName
}

func (typeFiltersConfig) TryMigrate(_ map[string]any) (result map[string]any, migrated bool) {
	return nil, false
}

func (a *app) loadTypeFiltersConfig() {
	a.ConfigRegister(&typeFiltersConfig{
		Version: typeFiltersConfigVersion,
	})
}

func (a *app) typeFiltersConfig() *typeFiltersConfig {
	if cfg, ok := a.ConfigFind(typeFiltersConfigName).(*typeFiltersConfig); ok {
		return cfg
	}
	log.Fatal().Msg("can't find typefilters config")
	return nil
}

// applyCustomFiltersFromConfig populates the pathsFilter with saved custom filter definitions.
// Called after a new environment (and fresh pathsFilter) is created.
func (a *app) applyCustomFiltersFromConfig() {
	for _, f := range a.typeFiltersConfig().CustomFilters {
		a.pathsFilter.AddCustomFilter(f.Name, f.Paths)
	}
}

// saveCustomFiltersToConfig syncs the current in-memory CustomFilters to the config.
func (a *app) saveCustomFiltersToConfig() {
	cfg := a.typeFiltersConfig()
	cfg.CustomFilters = nil
	for _, f := range a.pathsFilter.CustomFilters {
		cfg.CustomFilters = append(cfg.CustomFilters, savedCustomFilter{
			Name:  f.Name,
			Paths: f.Paths,
		})
	}
	a.configSaveV(cfg)
}
