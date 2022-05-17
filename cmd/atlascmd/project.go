// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package atlascmd

import (
	"fmt"
	"os"

	"ariga.io/atlas/schema/schemaspec"
	"ariga.io/atlas/schema/schemaspec/schemahcl"
)

const projectFileName = "atlas.hcl"

// projectFile represents an atlas.hcl file.
type projectFile struct {
	Envs []*Env `spec:"env"`
}

// MigrationDir represents the migration directory for the Env.
type MigrationDir struct {
	URL    string `spec:"url"`
	Format string `spec:"format"`
}

// Env represents an Atlas environment.
type Env struct {
	// Name for this environment.
	Name string `spec:"name,name"`

	// URL of the database.
	URL string `spec:"url"`

	// URL of the dev-database for this environment.
	// See: https://atlasgo.io/dev-database
	DevURL string `spec:"dev"`

	// Path to the file containing the desired schema of the environment.
	Source string `spec:"src"`

	// List of schemas in this database that are managed by Atlas.
	Schemas []string `spec:"schemas"`

	// Directory containing the migrations for the env.
	MigrationDir *MigrationDir `spec:"migration_dir"`

	schemaspec.DefaultExtension
}

var hclState = schemahcl.New(
	schemahcl.WithScopedEnums("env.migration_dir.format", formatAtlas, formatFlyway, formatLiquibase, formatGoose, formatGolangMigrate),
)

// LoadEnv reads the project file in path, and loads the environment
// with the provided name into env.
func LoadEnv(path string, name string) (*Env, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var project projectFile
	if err := hclState.Eval(b, &project, nil); err != nil {
		return nil, fmt.Errorf("error reading project file: %w", err)
	}
	projEnvs := make(map[string]*Env)
	for _, e := range project.Envs {
		if _, ok := projEnvs[e.Name]; ok {
			return nil, fmt.Errorf("duplicate environment name %q", e.Name)
		}
		if e.Name == "" {
			return nil, fmt.Errorf("all envs must have names on file %q", path)
		}
		if e.URL == "" {
			return nil, fmt.Errorf("no url set for env %q", e.Name)
		}
		if e.Source == "" {
			return nil, fmt.Errorf("no src set for env %q", e.Name)
		}
		projEnvs[e.Name] = e
	}
	selected, ok := projEnvs[name]
	if !ok {
		return nil, fmt.Errorf("env %q not defined in project file", name)
	}
	if selected.MigrationDir == nil {
		selected.MigrationDir = &MigrationDir{}
	}
	return selected, nil
}

func init() {
	schemaspec.Register("env", &Env{})
}