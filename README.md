![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/glennsarti/sentinel-parser/Go%20(Lint)?logo=github&label=Lint)

> [!IMPORTANT]
> This repository is **not** supported or endorsed by the authors of Sentinel (HashiCorp). Do not contact them if you find any issues with this parser!

---

# Sentinel Parser

The repository holds a parser for [HashiCorp Sentinel Configuration](https://developer.hashicorp.com/sentinel/docs/configuration) files, and a lexer and parser for [HashiCorp Sentinel Language](https://developer.hashicorp.com/sentinel/docs/language) files. The parsers support different versions of the language specifications.

> [!WARNING]
> This project is still a work in progress

# Intended usage of these Go modules

These Go modules are intended to be used by other tools to inspect Sentinel files e.g. development tools like;

* Static analysis tools (Linters etc.)

* Automation tools (Documentation generators)

* Text editor integrations ([Language Servers](https://microsoft.github.io/language-server-protocol/))

## What it shouldn't be used for

These Go modules are not intended to replace the [HashiCorp Sentinel CLI tool](https://developer.hashicorp.com/sentinel/docs/commands), for example; While not open-sourced, the parser in that tool would _probably_ be optimised to execute sentinel language which means it will terminate quickly on a parsing error. Whereas this parser is optimised to generate as much of an AST as possible, and be very forgiving of parsing errors.

# Example Usage

## Sentinel Configuration

```go
  import "github.com/glennsarti/sentinel-parser/sentinel_config/parser"

  func main() {
    // Create a new Sentinel Config parser using `sentinelVersion` language rules
    //
    // sentinelVersion can be 'latest' which will use the latest specification
    // the package knows, or a specific version like 'v0.25.0'. The complete list of
    // versions is found at `github.com/glennsarti/sentinel-parser/sentinel_config/features`
    p, err := parser.NewParser(sentinelVersion)
    if err != nil {
      return nil, err
    }

    content := []byte(`
      param "number1" {
        value = 42
      }
    `)

    // Parse a single file. Returns as much AST as possible and any parsing errors
    ast, diags := p.ParseFileSource(content, "sentinel.hcl")
  }

```

## Sentinel

Simple example

```go
  import "github.com/glennsarti/sentinel-parser/sentinel/parser"

  func main() {
    content := []byte(`
      main = rule { true }
    `)

    // Parse a single file. Returns as much AST as possible, any parsing errors and any terminal errors
    //
    // sentinelVersion can be 'latest' which will use the latest specification
    // the package knows, or a specific version like 'v0.25.0'. The complete list of
    // versions is found at `github.com/glennsarti/sentinel-parser/sentinel_config/features`
    ast, _, diags, err := parser.ParseFile(sentinelVersion, "example.sentinel", content)
  }
```

Advanced example

```go
  import (
    "github.com/glennsarti/sentinel-parser/sentinel/lexer"
    "github.com/glennsarti/sentinel-parser/sentinel/parser"
  )

  func main() {
    content := []byte(`
      main = rule { true }
    `)

    // Create a new lexer using the specific sentinelVersion specification
    lex, err := lexer.New(sentinelVersion, content)
    if err != nil {
      panic(err)
    }

    // Create a new parser using the specific sentinelVersion
    p, err := parser.New(sentinelVersion)
    if err != nil {
      panic(err)
    }

    // Parse the tokens from the lexer. The `nil` is to disable debug tracing.
    // Returns as much AST as possible, any parsing errors and any terminal errors
    ast, diags, err := p.ParseFile(lex, lex.Locator(), nil)
  }
```

# Language Specifications

These parsers were built according to the [language specifications](https://developer.hashicorp.com/sentinel/docs/language) as published by HashiCorp.

## Additional content used

The language specifications do not contain an exhaustive list of examples. To supplement these examples, the following files were used to inspect what "good" or "bad" examples could look like;

* The publicly published Sentinel policy packs in the [Terraform Registry](https://registry.terraform.io/browse/policies).

# Thanks

A huge thankyou to Thorsten Ball for publishing his book "Writing an Interpreter in Go" ([https://interpreterbook.com](https://interpreterbook.com)). This was instrumental in creating the lexer and parser in this project.

# TODO

There are still many thing to complete

* [ ] Generate Golang documentation

* [ ] Does it need more Sentinel language grammar checks

* [ ] Describe more about the generation scripts

* [ ] Actually do a 0.0.1 release

* [ ] What does a v1.0 release look like?
