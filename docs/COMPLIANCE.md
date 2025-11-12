# Documentation Compliance Guide

This guide ensures all documentation changes comply with Terraform Registry requirements and project standards.

## 🔄 Automated Validation Pipeline

### GitHub Actions Workflow

The `.github/workflows/documentation.yml` workflow automatically validates documentation on every pull request and push to main. It includes:

#### 1. Documentation Generation Check
- **Purpose**: Ensures generated docs are up-to-date
- **Runs**: On every change to docs, templates, examples, or provider code
- **Validates**:
  - Documentation can be generated without errors
  - Generated docs match the current code state
  - No manual edits to generated files

**Fix**: Run `make docs` or `cd tools && go generate`

#### 2. Example Validation
- **Purpose**: Ensures example files meet quality standards
- **Validates**:
  - Copyright headers present (`Copyright (c) 2025 Michael Villani`)
  - SPDX license identifier (`SPDX-License-Identifier: MPL-2.0`)
  - Terraform formatting (`terraform fmt -check`)
  - Valid Terraform syntax (`terraform validate`)
  - No hardcoded credentials or sensitive data

**Fix**: 
```bash
# Add headers to examples
# Run formatting
terraform fmt -recursive examples/
```

#### 3. Template Validation
- **Purpose**: Ensures templates follow required structure
- **Validates**:
  - Required template files exist
  - Templates contain `{{ .SchemaMarkdown }}` placeholder
  - Templates have proper YAML front matter
  - Required placeholders present (`{{.Type}}`, `{{.Name}}`)

**Fix**: Review `templates/README.md` for template requirements

#### 4. Documentation Structure
- **Purpose**: Validates generated documentation structure
- **Validates**:
  - Required files exist (`docs/index.md`)
  - YAML front matter in all docs
  - Schema sections present in resources/data sources
  - Example usage sections present
  - No broken internal links

**Fix**: Run documentation generation and review output

## 🛠️ Local Development Tools

### Make Targets

We provide several Make targets for documentation tasks:

```bash
# Generate documentation
make docs

# Check if documentation is up-to-date
make docs-check

# Validate example formatting
make validate-examples

# Run all validation checks
make validate-docs

# Full build with documentation
make
```

### Pre-commit Hooks

Install pre-commit hooks to automatically validate documentation before committing:

```bash
# Install pre-commit (if not already installed)
pip install pre-commit

# Install the hooks
pre-commit install

# Run manually on all files
pre-commit run --all-files
```

The pre-commit hooks will:
- ✅ Check copyright headers in examples
- ✅ Format Terraform examples
- ✅ Validate template structure
- ✅ Auto-generate documentation
- ✅ Ensure docs are in sync

## 📋 Manual Validation Checklist

Before submitting a pull request with documentation changes:

### For New Resources/Data Sources

- [ ] Template file created in `templates/resources/` or `templates/data-sources/`
- [ ] Template includes all required sections:
  - [ ] YAML front matter
  - [ ] Description
  - [ ] Important notes/warnings (if applicable)
  - [ ] Multiple usage examples
  - [ ] `{{ .SchemaMarkdown }}` placeholder
  - [ ] Import instructions (resources only)
  - [ ] Timeouts documentation (if applicable)
  - [ ] Best practices section
  - [ ] Common issues section
  - [ ] Related resources links
- [ ] Example file created in `examples/resources/` or `examples/data-sources/`
- [ ] Example includes:
  - [ ] Copyright header
  - [ ] SPDX license identifier
  - [ ] Complete, working configuration
  - [ ] Helpful comments
  - [ ] No hardcoded credentials
- [ ] Documentation generated: `make docs`
- [ ] Examples formatted: `terraform fmt -recursive examples/`
- [ ] Generated docs reviewed for accuracy

### For Documentation Updates

- [ ] Changes made to templates, not generated docs
- [ ] Examples updated if schema changed
- [ ] Documentation regenerated after changes
- [ ] No manual edits to `docs/` directory
- [ ] All validation checks pass: `make validate-docs`

### For Example Updates

- [ ] Examples have copyright headers
- [ ] Examples are formatted: `terraform fmt -recursive examples/`
- [ ] Examples are syntactically valid
- [ ] No sensitive data or credentials
- [ ] Examples demonstrate realistic use cases
- [ ] Documentation regenerated if examples referenced in templates

## 🚨 Common Issues and Solutions

### Issue: Documentation out of sync

**Error**: `Documentation is out of date`

**Solution**:
```bash
cd tools
go generate
git add docs/
git commit -m "Update generated documentation"
```

### Issue: Example formatting

**Error**: `Example files are not properly formatted`

**Solution**:
```bash
terraform fmt -recursive examples/
git add examples/
git commit -m "Format example files"
```

### Issue: Missing copyright headers

**Error**: `Missing copyright header in examples/...`

**Solution**: Add to the top of each `.tf` file:
```terraform
# Copyright (c) 2025 Michael Villani
# SPDX-License-Identifier: MPL-2.0
```

### Issue: Missing template placeholders

**Error**: `Missing {{ .SchemaMarkdown }} placeholder`

**Solution**: Add to your template where the schema should appear:
```markdown
{{ .SchemaMarkdown | trimspace }}
```

### Issue: Invalid YAML front matter

**Error**: `Missing YAML front matter`

**Solution**: Ensure template starts with:
```yaml
---
page_title: "{{.Type}} {{.Name}} - {{.ProviderName}}"
subcategory: ""
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---
```

## 📊 Documentation Quality Metrics

The pipeline reports the following metrics on each PR:

| Metric | Description |
|--------|-------------|
| Templates | Number of template files |
| Generated Docs | Number of generated documentation files |
| Example Files | Number of example `.tf` files |
| Resources Documented | Number of resource documentation pages |
| Data Sources Documented | Number of data source documentation pages |
| Total Doc Lines | Total lines of documentation |
| Total Example Lines | Total lines of example code |

## 🎯 Validation Pipeline Flow

```
┌─────────────────────────────────────────────────────────────┐
│                     Pull Request Created                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
    ┌────────────────────────────────────────────────────┐
    │        Documentation Validation Workflow            │
    └────────────────────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┬──────────────┐
        │                │                │              │
        ▼                ▼                ▼              ▼
┌───────────────┐ ┌──────────────┐ ┌──────────┐ ┌─────────────┐
│ Documentation │ │   Example    │ │ Template │ │Documentation│
│ Generation    │ │ Validation   │ │Validation│ │  Structure  │
│     Check     │ │              │ │          │ │             │
└───────┬───────┘ └──────┬───────┘ └────┬─────┘ └──────┬──────┘
        │                │              │              │
        └────────────────┴──────────────┴──────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │  All Checks Pass?    │
              └──────────┬───────────┘
                         │
                ┌────────┴────────┐
                │                 │
                ▼                 ▼
             ┌─────┐          ┌──────┐
             │ YES │          │  NO  │
             └──┬──┘          └───┬──┘
                │                 │
                ▼                 ▼
         ┌───────────┐      ┌──────────┐
         │ Report ✅ │      │Report ❌ │
         │ Posted    │      │Posted    │
         └───────────┘      └──────────┘
```

## 🔐 Required Checks for Merge

All of the following must pass before merging:

1. ✅ Documentation Generation Check
2. ✅ Example Validation
3. ✅ Template Validation
4. ✅ Documentation Structure
5. ✅ All other CI/CD checks (tests, linting, etc.)

## 📚 Additional Resources

- [Terraform Registry Documentation Requirements](https://developer.hashicorp.com/terraform/registry/providers/docs)
- [terraform-plugin-docs Tool](https://github.com/hashicorp/terraform-plugin-docs)
- [Project Documentation Templates](../templates/README.md)
- [Documentation Validation Checklist](../DOCUMENTATION.md)
- [Provider Instructions](../.github/copilot-instructions.md#documentation-requirements)

## 🤝 Getting Help

If you encounter issues with documentation validation:

1. Check this guide for common issues
2. Review the [templates/README.md](../templates/README.md) file
3. Look at existing templates for examples
4. Check the GitHub Actions workflow logs for detailed error messages
5. Run validation locally: `make validate-docs`

## 📝 Workflow Configuration

The documentation validation workflow is configured in:
- **File**: `.github/workflows/documentation.yml`
- **Triggers**: Pull requests and pushes affecting documentation
- **Timeout**: 10 minutes per job (5 minutes for simple checks)
- **Permissions**: Read repository, write to pull request comments

### Workflow Jobs

1. **documentation-check** (10 min)
   - Generates documentation
   - Compares with committed files
   - Reports differences

2. **example-validation** (10 min)
   - Checks copyright headers
   - Validates formatting
   - Checks syntax
   - Scans for credentials

3. **template-validation** (5 min)
   - Verifies required files
   - Validates placeholders
   - Checks front matter

4. **documentation-structure** (5 min)
   - Validates file structure
   - Checks schema sections
   - Verifies example usage
   - Scans for broken links

5. **documentation-report** (Always runs)
   - Generates statistics
   - Posts summary to PR

## 🔄 Continuous Improvement

This validation pipeline is continuously improved based on:
- Terraform Registry requirement updates
- Community feedback
- Common documentation issues
- Best practices evolution

To suggest improvements, please open an issue or pull request.
