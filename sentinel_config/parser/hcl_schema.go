package parser

import (
	"github.com/glennsarti/sentinel-parser/features"
	"github.com/hashicorp/hcl/v2"
)

func buildHclSchema(body hcl.Body, sentinelVersion string) (*hcl.BodySchema, error) {
	// Build from the common schema
	schema := hcl.BodySchema{
		Blocks: make([]hcl.BlockHeaderSchema, len(ConfigCommonBodySchema.Blocks)),
	}
	for idx, bhs := range ConfigCommonBodySchema.Blocks {
		schema.Blocks[idx] = cloneBlockHeaderSchema(bhs)
	}

	// Determine the import block version
	schema.Blocks = append(schema.Blocks, buildImportBlockSchema(body, sentinelVersion))

	return &schema, nil
}

func buildImportBlockSchema(body hcl.Body, sentinelVersion string) hcl.BlockHeaderSchema {
	if features.UnsupportedVersion(sentinelVersion, features.V2ImportBlockMinimumVersion) {
		return v1ImportBlockHeaderSchema
	}

	// Inspect the file for V1 block schema
	c, _, d := body.PartialContent(V1ImportBodySchema)
	if !d.HasErrors() && len(c.Blocks) > 0 {
		return v1ImportBlockHeaderSchema
	}

	return v2ImportBlockHeaderSchema
}

func cloneBlockHeaderSchema(source hcl.BlockHeaderSchema) hcl.BlockHeaderSchema {
	bhs := hcl.BlockHeaderSchema{
		Type:       source.Type,
		LabelNames: make([]string, len(source.LabelNames)),
	}
	copy(bhs.LabelNames, source.LabelNames)

	return bhs
}

var v2ImportBlockHeaderSchema = hcl.BlockHeaderSchema{
	Type:       "import",
	LabelNames: []string{"kind", "name"},
}

var v1ImportBlockHeaderSchema = hcl.BlockHeaderSchema{
	Type:       "import",
	LabelNames: []string{"name"},
}

var V1ImportBodySchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{v1ImportBlockHeaderSchema},
}

var EmptyBodySchema = &hcl.BodySchema{}

var ConfigCommonBodySchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "sentinel",
		},
		{
			Type:       "policy",
			LabelNames: []string{"name"},
		},
		{
			Type:       "module",
			LabelNames: []string{"name"},
		},
		{
			Type:       "mock",
			LabelNames: []string{"name"},
		},
		{
			Type:       "global",
			LabelNames: []string{"name"},
		},
		{
			Type:       "param",
			LabelNames: []string{"name"},
		},
		{
			Type: "test",
		},
	},
}

var SentinelBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "features",
			Required: false,
		},
	},
}

var PolicyBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "source",
			Required: false,
		},
		{
			Name:     "enforcement_level",
			Required: false,
		},
		{
			Name:     "params",
			Required: false,
		},
	},
}

var MockBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "data",
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "module",
		},
	},
}

var MockModuleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "source",
			Required: false,
		},
	},
}

var GlobalBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "value",
			Required: false,
		},
	},
}

var ParamBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "value",
			Required: false,
		},
	},
}

var TestBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "rules",
			Required: false,
		},
	},
}

var v1ImportModuleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "source",
			Required: false,
		},
	},
}

var v1ImportPluginBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "path",
			Required: false,
		},
		{
			Name:     "args",
			Required: false,
		},
		{
			Name:     "env",
			Required: false,
		},
		{
			Name:     "config",
			Required: false,
		},
	},
}

var v2ImportModuleBlockSchema = v1ImportModuleBlockSchema

var v2ImportPluginBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "source",
			Required: false,
		},
		{
			Name:     "args",
			Required: false,
		},
		{
			Name:     "env",
			Required: false,
		},
		{
			Name:     "config",
			Required: false,
		},
	},
}

var v2ImportStaticBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "source",
			Required: false,
		},
		{
			Name:     "format",
			Required: false,
		},
	},
}
