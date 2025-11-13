# Documentation Templates

This directory contains templates for generating the provider documentation using [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs).

## Overview

The Terraform provider documentation is **generated automatically** from these templates. Do not edit files in the `docs/` directory directly - they will be overwritten when documentation is regenerated.

## Directory Structure

```
templates/
├── index.md.tmpl                    # Provider overview and configuration
├── resources/                       # Resource templates
│   └── environment.md.tmpl         # bcadmincenter_environment resource
└── data-sources/                   # Data source templates
    └── available_applications.md.tmpl  # bcadmincenter_available_applications data source
```

## Generating Documentation

To regenerate the documentation after making changes:

```bash
cd tools
go generate
```

This will:
1. Extract schema information from the provider code
2. Process the templates in this directory
3. Include example files from `examples/`
4. Generate final documentation in `docs/`

## Template Syntax

Templates use Go's `text/template` syntax with special functions provided by terraform-plugin-docs:

### Available Placeholders

- `{{ .SchemaMarkdown }}` - Auto-generated schema documentation
- `{{ .Description }}` - Resource/data source description from code
- `{{ .Type }}` - "Resource" or "Data Source"
- `{{ .Name }}` - Resource/data source name (e.g., "bcadmincenter_environment")
- `{{ .ProviderName }}` - Provider name ("bcadmincenter")

### Including Examples

Use the `tffile` function to include example Terraform files:

```markdown
{{tffile "examples/resources/bc_admin_center_environment/resource.tf"}}
```

This will:
- Read the specified file
- Format it as a Terraform code block
- Include it in the generated documentation

## Writing Templates

### 1. Front Matter

Every template must start with YAML front matter:

```yaml
---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---
```

### 2. Headings

Use standard Markdown headings:

```markdown
# {{.Type}} ({{.Name}})

## Example Usage

### Basic Example
```

### 3. Callouts

Use Terraform-specific callout syntax:

```markdown
~> **Warning:** This is a warning message

-> **Note:** This is a note message
```

### 4. Code Blocks

Use fenced code blocks with language specification:

````markdown
```terraform
resource "bcadmincenter_environment" "example" {
  name = "production"
}
```

```bash
terraform import bcadmincenter_environment.example tenant-id/BusinessCentral/production
```
````

## Template Requirements

Each template should include:

### Provider Template (index.md.tmpl)

- [ ] Clear provider description
- [ ] Authentication methods with examples
- [ ] Required permissions
- [ ] Environment variables table
- [ ] Multiple usage examples
- [ ] `{{ .SchemaMarkdown }}` placeholder
- [ ] Links to additional resources

### Resource Templates

- [ ] Clear resource description
- [ ] Important warnings (using `~>`)
- [ ] Multiple usage examples
- [ ] Import instructions with examples
- [ ] Timeouts documentation (if applicable)
- [ ] `{{ .SchemaMarkdown }}` placeholder
- [ ] Best practices section
- [ ] Common issues section

### Data Source Templates

- [ ] Clear data source description
- [ ] Usage examples with common patterns
- [ ] Use cases with resources
- [ ] `{{ .SchemaMarkdown }}` placeholder
- [ ] Attribute reference explanation

## Example Files

Templates reference example files from the `examples/` directory:

```
examples/
├── provider/
│   └── provider.tf                 # Provider configuration
├── resources/
│   └── bc_admin_center_environment/
│       └── resource.tf             # Basic resource example
└── data-sources/
    └── bc_admin_center_available_applications/
        └── data-source.tf          # Basic data source example
```

### Example File Requirements

All example files must:
- Include copyright headers
- Be complete, working configurations
- Use realistic but safe values
- Include helpful comments
- Be formatted with `terraform fmt`

## Best Practices

1. **Keep templates focused** - Each template should cover one resource/data source
2. **Use consistent formatting** - Follow Terraform documentation conventions
3. **Include multiple examples** - Show basic and advanced usage
4. **Link to related docs** - Reference Microsoft documentation where appropriate
5. **Update examples** - Keep example files synchronized with templates
6. **Test generation** - Always regenerate and review docs after template changes

## Markdown Formatting

### Line Length

Use semantic line breaks (break at sentence boundaries) for better git diffs:

```markdown
This is a long sentence that explains something important.
It breaks at the sentence boundary for better version control.
Another sentence starts on a new line.
```

### Tables

Use standard Markdown tables:

```markdown
| Variable | Description |
|----------|-------------|
| `AZURE_CLIENT_ID` | The client ID |
| `AZURE_TENANT_ID` | The tenant ID |
```

### Links

Use descriptive link text:

```markdown
See the [Business Central Admin Center API documentation](https://learn.microsoft.com/...)
```

## Troubleshooting

### Schema Not Showing

If `{{ .SchemaMarkdown }}` doesn't render:
- Ensure the provider builds successfully
- Check that resource/data source is registered in provider code
- Verify template syntax is correct

### Examples Not Including

If `{{tffile ...}}` doesn't work:
- Verify the file path is correct (relative to repo root)
- Check that the example file exists
- Ensure the file has proper formatting

### Generation Fails

If `go generate` fails:
- Check template syntax for errors
- Verify all referenced example files exist
- Review the error message for specific issues
- Ensure provider builds successfully

## Adding New Documentation

When adding a new resource or data source:

1. Create the template file:
   ```bash
   # For resources
   touch templates/resources/new_resource.md.tmpl
   
   # For data sources
   touch templates/data-sources/new_data_source.md.tmpl
   ```

2. Create the example file:
   ```bash
   # For resources
   mkdir -p examples/resources/bc_admin_center_new_resource
   touch examples/resources/bc_admin_center_new_resource/resource.tf
   
   # For data sources
   mkdir -p examples/data-sources/bc_admin_center_new_data_source
   touch examples/data-sources/bc_admin_center_new_data_source/data-source.tf
   ```

3. Write the template following the guidelines above

4. Add copyright headers to example files

5. Generate documentation:
   ```bash
   cd tools
   go generate
   ```

6. Review generated documentation in `docs/`

7. Commit both templates and generated docs

## Resources

- [terraform-plugin-docs Documentation](https://github.com/hashicorp/terraform-plugin-docs)
- [Terraform Registry Provider Documentation](https://developer.hashicorp.com/terraform/registry/providers/docs)
- [Terraform Documentation Style Guide](https://www.terraform.io/docs/extend/best-practices/writing-documentation.html)
