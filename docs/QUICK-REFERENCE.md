# Documentation Quick Reference

Quick commands and checklist for working with provider documentation.

## 🚀 Quick Commands

```bash
# Generate documentation
make docs
# or
cd tools && go generate

# Check documentation is up-to-date
make docs-check

# Validate examples
make validate-examples

# Run all validations
make validate-docs

# Format examples
terraform fmt -recursive examples/

# Install pre-commit hooks
pre-commit install

# Run pre-commit manually
pre-commit run --all-files
```

## ✅ Quick Checklist

### Before Committing Changes

- [ ] Run `make docs` if templates or code changed
- [ ] Run `terraform fmt -recursive examples/` if examples changed
- [ ] Verify copyright headers in new example files
- [ ] No hardcoded credentials in examples
- [ ] Templates have `{{ .SchemaMarkdown }}` placeholder
- [ ] Run `make validate-docs` to check everything

### Creating New Resource/Data Source

1. Create template: `templates/resources/your_resource.md.tmpl`
2. Create example: `examples/resources/bc_admin_center_your_resource/resource.tf`
3. Add copyright header to example
4. Run `make docs`
5. Review `docs/resources/your_resource.md`
6. Commit all changes (template + example + generated docs)

## 📁 File Locations

| Type | Location | Edit? |
|------|----------|-------|
| Provider docs template | `templates/index.md.tmpl` | ✅ Yes |
| Resource templates | `templates/resources/*.md.tmpl` | ✅ Yes |
| Data source templates | `templates/data-sources/*.md.tmpl` | ✅ Yes |
| Generated docs | `docs/**/*.md` | ❌ No - Auto-generated |
| Provider examples | `examples/provider/provider.tf` | ✅ Yes |
| Resource examples | `examples/resources/*/resource.tf` | ✅ Yes |
| Data source examples | `examples/data-sources/*/data-source.tf` | ✅ Yes |

## 🔧 Template Placeholders

Must include in all templates:
- `{{ .SchemaMarkdown }}` - Generates the schema documentation
- `{{.Type}}` - Resource or Data Source
- `{{.Name}}` - Resource/data source name
- `{{.ProviderName}}` - Provider name (bcadmincenter)

Example inclusion:
- `{{tffile "examples/path/to/file.tf"}}` - Includes example file

## 📋 Required Template Sections

### All Templates
```markdown
---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Type}} ({{.Name}})

{{ .Description | trimspace }}

## Example Usage

{{ .SchemaMarkdown | trimspace }}
```

### Resources Only
- Import instructions with example
- Timeouts (if applicable)

### Additional Sections (Recommended)
- Important Notes with warnings (`~>`)
- Best Practices
- Common Issues
- Related Resources

## 🚨 Common Errors

| Error | Fix |
|-------|-----|
| Documentation out of date | `make docs` |
| Examples not formatted | `terraform fmt -recursive examples/` |
| Missing copyright header | Add to top of `.tf` file |
| Missing schema placeholder | Add `{{ .SchemaMarkdown }}` to template |
| Invalid front matter | Check YAML syntax at top of template |

## 💡 Tips

- **Don't edit `docs/` directly** - Edit templates instead
- **Always regenerate** after template changes
- **Test locally** before pushing: `make validate-docs`
- **Use pre-commit hooks** to catch issues early
- **Copy existing templates** as a starting point
- **Check CI/CD** for detailed error messages

## 📖 Full Documentation

- Complete guide: [docs/COMPLIANCE.md](COMPLIANCE.md)
- Template guide: [templates/README.md](../templates/README.md)
- Validation checklist: [DOCUMENTATION.md](../DOCUMENTATION.md)
- Instructions: [.github/copilot-instructions.md](../.github/copilot-instructions.md)
