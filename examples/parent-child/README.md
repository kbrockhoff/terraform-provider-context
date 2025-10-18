# Parent-Child Context Example

This example demonstrates how to use parent and child contexts to manage configuration for a multi-component application stack.

## Concept

When building a Terraform stack with multiple components (API, database, cache, etc.), many configuration values are shared across all components:
- Namespace and environment
- Cost center and ownership
- Project management IDs
- Common tags

The parent-child context pattern allows you to:
1. Define common values once in a **parent context**
2. Have **child contexts** inherit those values
3. Override only what's different for each component

## Example Structure

```
Payment Stack (Parent)
├── API (Child) - inherits all, adds component-specific tags
├── Database (Child) - inherits all, overrides sensitivity
└── Cache (Child) - inherits all, overrides availability
```

## Benefits

1. **DRY (Don't Repeat Yourself)**: Common values defined once
2. **Consistency**: All components share the same namespace, environment, cost center, etc.
3. **Flexibility**: Each component can override what's different
4. **Maintainability**: Change parent context to update all children

## Usage

The parent context establishes the baseline:

```hcl
data "brockhoff_context" "parent" {
  namespace    = "platform"
  environment  = "prod"
  cost_center  = "engineering"
  # ... other common settings
}
```

Child contexts inherit from the parent and override specific fields:

```hcl
data "brockhoff_context" "api" {
  # Inherit parent values
  context = data.brockhoff_context.parent.context_output

  # Override only what's different
  name = "payment-api"
  additional_tags = {
    component = "api"
  }
}
```

## Merge Logic

Values are resolved in this order:
1. **Defaults**: Built-in defaults (e.g., availability = "preemptable")
2. **Parent Context**: Values from `context` input
3. **Individual Inputs**: Explicitly set values (highest priority)

Example:
- Parent sets `cost_center = "engineering"`
- Child doesn't set `cost_center`
- Result: Child inherits `cost_center = "engineering"`

But if child sets `cost_center = "ops"`, that overrides the parent value.

## Output

Each context produces:
- `name_prefix`: Computed resource name
- `tags`: Merged tags map
- `context_output`: Resolved values for use by child contexts

The `context_output` contains all resolved configuration values, making it easy to pass to child contexts.

## Testing

```bash
terraform init
terraform plan
```

You'll see:
- Parent: `platform-prod`
- API: `platform-payment-api-prod` (inherits namespace/environment, adds name)
- Database: `platform-payment-db-prod` (inherits namespace/environment, adds name)
- Cache: `platform-payment-cache-prod` (inherits namespace/environment, adds name)

All components share the parent's cost center, ownership, and base tags, but each adds its own component-specific configuration.
