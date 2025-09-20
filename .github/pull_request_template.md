# Pull Request

## Description

<!-- Provide a brief description of the changes in this PR -->

## Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🔧 Refactoring (no functional changes, no API changes)
- [ ] ⚡ Performance improvement
- [ ] 🧪 Test updates
- [ ] 🔨 Build/CI changes
- [ ] 📦 Dependency updates

## Related Issues

<!-- Link to any related issues using "Fixes #issue_number" or "Related to #issue_number" -->

- Fixes #
- Related to #

## Changes Made

<!-- Provide a detailed list of changes made in this PR -->

- 
- 
- 

## Testing

<!-- Describe the tests you ran to verify your changes -->

### Test Commands Run

- [ ] `make test` - Unit tests pass
- [ ] `make lint` - Linting passes  
- [ ] `make build` - Build succeeds
- [ ] `make test-examples` - Example validation passes
- [ ] Manual testing performed

### Test Cases Covered

<!-- List specific test cases or scenarios tested -->

- [ ] 
- [ ] 
- [ ] 

## Provider Compatibility

<!-- Mark all cloud providers tested -->

- [ ] AWS
- [ ] Azure  
- [ ] GCP
- [ ] Default/DC
- [ ] Other cloud providers (specify): ___________

## Breaking Changes

<!-- If this is a breaking change, describe the impact and migration path -->

### Impact

<!-- Describe what will break -->

### Migration Path

<!-- Describe how users can migrate to the new version -->

## Documentation

<!-- Mark all that apply -->

- [ ] Code comments updated
- [ ] README updated
- [ ] Example configurations updated
- [ ] Provider documentation needs regeneration
- [ ] No documentation changes needed

## Checklist

<!-- Ensure all items are completed before requesting review -->

### Code Quality

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings or errors
- [ ] I have added error handling where appropriate

### Testing

- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have tested my changes with example configurations
- [ ] I have verified backward compatibility (if applicable)

### Security

- [ ] My changes do not introduce security vulnerabilities
- [ ] I have not hardcoded sensitive information
- [ ] Input validation is properly implemented
- [ ] Error messages do not leak sensitive information

### Performance

- [ ] My changes do not negatively impact performance
- [ ] I have considered memory usage implications
- [ ] External calls are minimized and cached where appropriate

## Reviewer Guidelines

<!-- Information for reviewers -->

### Focus Areas

<!-- Highlight specific areas that need extra attention during review -->

- [ ] Business logic correctness
- [ ] Error handling
- [ ] Performance implications
- [ ] Security considerations
- [ ] API compatibility
- [ ] Test coverage

### Testing Instructions

<!-- Specific instructions for reviewers to test the changes -->

1. 
2. 
3. 

## Additional Context

<!-- Add any other context, screenshots, or information that would be helpful for reviewers -->

## Release Notes

<!-- If this change should be mentioned in release notes, provide a brief description -->

```
<!-- Example:
- feat: Add support for ephemeral environment auto-deletion
- fix: Resolve issue with tag truncation for GCP provider
- breaking: Change provider configuration structure (see migration guide)
-->
```

---

**Note**: This PR will be automatically tested against multiple Terraform versions and cloud provider configurations. Please ensure all tests pass before requesting review.