package cnb

import (
	"encoding/json"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const buildMetadataLabel = "io.buildpacks.build.metadata"

type BuildMetadata struct {
	Processes  []Process   `json:"processes"`
	Buildpacks []Buildpack `json:"buildpacks"`
	BOM        []BOMEntry  `json:"bom"`
}

type Buildpack struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type BOMEntry struct {
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Metadata  map[string]interface{} `json:"metadata"`
	Buildpack Buildpack
}

type Process struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Direct  bool     `json:"direct"`
}

func (m *BuildMetadata) FindBOM(name string) BOMEntry {
	for _, entry := range m.BOM {
		if entry.Name == name {
			return entry
		}
	}
	return BOMEntry{}
}

func ParseBuildMetadata(img v1.Image) (BuildMetadata, error) {
	cfg, err := img.ConfigFile()
	if err != nil {
		return BuildMetadata{}, err
	}
	label, ok := cfg.Config.Labels[buildMetadataLabel]
	if !ok {
		return BuildMetadata{}, nil
	}
	var md BuildMetadata
	if err := json.Unmarshal([]byte(label), &md); err != nil {
		return BuildMetadata{}, err
	}
	return md, nil
}
