# Generation Tools

## AST Generation

Many parts of the parser are boiler-plate code in particular the serialisation/deserialisation of ASTs. To help with this these tools can generate the code based on a (hopefully?) simple schema document in JSON

## Usage

`node <filename>`

### Schema `sentinel_ast_schema.json`

| File | Description |
| ---- | ---- |
| `generate_sentinel_json.js` | Creates the JSON serializer/deserializer for the Sentinel AST. |
| `generate_sentinel_ast_walker.js` | Creates a walker for the Sentinel AST. |
| `generate_sentinel_ast.js` | Creates the AST objects. |

### Schema `sentinel_config_ast_schema.json`

| File | Description |
| ---- | ---- |
| `generate_sentinel_config_json.js` | Creates the JSON serializer/deserializer for the Sentinel Config AST. |
| `generate_sentinel_config_ast.js` | Creates the AST objects. |

## Version Generation

The `generate_sentinel_versions.js` queries the Sentinel Releases website and extracts all of the versions of the Sentinel CLI.
