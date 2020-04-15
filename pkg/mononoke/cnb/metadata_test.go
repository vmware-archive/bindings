package cnb

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
)

func TestParseBuildMetadata(t *testing.T) {
	img := &fake.FakeImage{
		ConfigFileStub: func() (*v1.ConfigFile, error) {
			return &v1.ConfigFile{
				Config: v1.Config{
					Labels: map[string]string{
						"io.buildpacks.build.metadata": testLabel,
					},
				},
			}, nil
		},
	}
	md, err := ParseBuildMetadata(img)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if diff := cmp.Diff(md.Processes, []Process{
		{
			Type:    "process-type-1",
			Command: "/some/command to run with bash",
			Args:    nil,
			Direct:  false,
		},
		{
			Type:    "process-type-2",
			Command: "/other/command",
			Args:    []string{"to", "directly", "execute"},
			Direct:  true,
		},
	}); diff != "" {
		t.Fatalf("processes (-expected, +actual) = %v", diff)
	}

	if diff := cmp.Diff(md.Buildpacks, []Buildpack{
		{
			ID:      "com.example.buildpack.1",
			Version: "v1.2.3",
		},
		{
			ID:      "com.example.buildpack.2",
			Version: "v3.2.1",
		},
	}); diff != "" {
		t.Fatalf("buildpacks (-expected, +actual) = %v", diff)
	}

	if diff := cmp.Diff(md.BOM, []BOMEntry{
		{
			Name:    "entry-1",
			Version: "0.0.1",
			Metadata: map[string]interface{}{
				"structured": []interface{}{
					interface{}(map[string]interface{}{
						"k1": "v1",
						"k2": "v2",
					}),
				},
				"k3": "v3",
			},
			Buildpack: Buildpack{
				ID:      "com.example.buildpack.1",
				Version: "v1.2.3",
			},
		},
		{
			Name:     "entry-2",
			Version:  "0.0.2",
			Metadata: map[string]interface{}{},
			Buildpack: Buildpack{
				ID:      "com.example.buildpack.2",
				Version: "v3.2.1",
			},
		},
	}); diff != "" {
		t.Fatalf("buildpacks (-expected, +actual) = %v", diff)
	}
}

var testLabel = `
{
  "processes": [
    {
      "type": "process-type-1",
      "command": "/some/command to run with bash",
      "args": null,
      "direct": false
    },
    {
      "type": "process-type-2",
      "command": "/other/command",
      "args": ["to", "directly", "execute"],
      "direct": true
    }
  ],
  "buildpacks": [
    {
      "id": "com.example.buildpack.1",
      "version": "v1.2.3"
    },
    {
      "id": "com.example.buildpack.2",
      "version": "v3.2.1"
    }
  ],
  "bom": [
    {
      "name": "entry-1",
      "version": "0.0.1",
      "metadata": {
        "structured": [
          {
            "k1": "v1",
            "k2": "v2"
          }
        ],
        "k3": "v3"
      },
      "buildpack": {
        "id": "com.example.buildpack.1",
        "version": "v1.2.3"
      }
    },
    {
      "name": "entry-2",
      "version": "0.0.2",
      "metadata": {},
      "buildpack": {
        "id": "com.example.buildpack.2",
        "version": "v3.2.1"
      }
    }
  ]
}
`
