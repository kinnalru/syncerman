# PLAN_7: Refactoring Internal Packages

## Overview
Template plan for refactoring internal packages. Each package has its own milestone with common refactoring tasks.

## Packages to Refactor

1. `internal/cmd`
2. `internal/config`
3. `internal/errors`
4. `internal/logger`
5. `internal/rclone`
6. `internal/sync`
7. `internal/version`

---

## Common Refactoring Tasks for Each Package

- [ ] **Logger Consolidation**
  - Check if package has its own Logger implementation
  - If yes, move it to `internal/logger` package
  - Import Logger from `internal/logger` instead
  - Remove local Logger implementation

- [ ] **Remove Unused Code**
  - Identify all functions and types in the package
  - Check usage across:
    - Package itself
    - Package tests
    - ALL other packages in project
  - Remove unused functions
  - Remove unused types
  - Update imports

- [ ] **Remove Unused Files**
  - List all files in package directory
  - Check if each file is imported or used elsewhere
  - Remove unused files

- [ ] **Verify Package Specification**
  - Read specification file in `specs/PACKAGE_<NAME>.md` (if exists)
  - Compare current implementation with specification
  - Add missing features/methods
  - Remove extra features/methods not in spec
  - Ensure API matches specification

- [ ] **Simplify Code**
  - Review for complex logic that can be simplified
  - Break down large functions
  - Reduce nesting levels
  - Improve readability
  - Remove code duplication

- [ ] **Apply DRY Principle**
  - Find repeated code patterns
  - Extract common logic into reusable functions
  - Use existing helpers from other packages when appropriate
  - Ensure single responsibility for each function

- [ ] **Check Tests**
  - each file in pacakge must has `*_test.go` test implementation

---

## Milestone 1: Refactor `internal/errors`

### Tasks
- [ ] Apply common refactoring tasks (see above)
- [ ] Review error types
- [ ] Consolidate error handling patterns
- [ ] Improve error wrapping
- [ ] Ensure consistent error formatting

### Notes
- Package contains custom error types and utilities
- No specification file found
- [ ] Specify additional errors-specific tasks here

---

## Milestone 2: Refactor `internal/logger`

### Tasks
- [ ] Apply common refactoring tasks (see above)
- [ ] Review logging levels
- [ ] Consolidate logging formats
- [ ] Review color usage
- [ ] Review semantic logging

### Notes
- Package provides structured logging
- Other packages MUST NOT have local Logger implementations
- No specification file found
- [ ] Specify additional logger-specific tasks here

---

## Milestone 3: Refactor `internal/version`

### Tasks
- [ ] Apply common refactoring tasks (see above)
- [ ] Review version information handling
- [ ] Check if package is still needed
- [ ] Consider moving version info to main or config

### Notes
- Package likely contains version/build information
- No specification file found
- [ ] Specify additional version-specific tasks here

---

## Milestone 4: Refactor `internal/config`

### Tasks
- [ ] Apply common refactoring tasks (see above)
- [ ] Review YAML parsing logic
- [ ] Simplify validation rules
- [ ] Consolidate configuration loading
- [ ] Improve error messages

### Notes
- Package handles configuration loading, parsing, and validation
- No specification file found
- [ ] Specify additional config-specific tasks here

---

## Milestone 5: Refactor `internal/rclone`

### Tasks
- [ ] Apply common refactoring tasks (see above)
- [ ] **Verify Package Specification**
  - Read `specs/PACKAGE_RCLONE.md`
  - Ensure all required methods exist
  - Ensure API matches specification exactly
- [ ] Review rclone command building logic
- [ ] Consolidate error handling

### Notes
- Package handles rclone command execution and verification
- Specification document: `specs/PACKAGE_RCLONE.md`
- Critical component - all sync operations depend on this package
- [ ] Specify additional rclone-specific tasks here

---

## Milestone 6: Refactor `internal/sync`

### Tasks
- [ ] Apply common refactoring tasks (see above)
- [ ] **Verify Package Specification**
  - Read `specs/PACKAGE_SYNC.md`
  - Ensure all required methods exist
  - Ensure API matches specification exactly
- [ ] Review sync orchestration logic
- [ ] Simplify sequential processing
- [ ] Improve first-run detection

### Notes
- Package contains core sync logic and orchestration
- Specification document: `specs/PACKAGE_SYNC.md`
- Handles sequential sync processing with error handling
- Implements order preservation from YAML configuration
- [ ] Specify additional sync-specific tasks here

---

## Milestone 7: Refactor `internal/cmd`

### Tasks
- [ ] Apply common refactoring tasks (see above)
- [ ] Review CLI command structure
- [ ] Consolidate flag handling logic
- [ ] Review and improve command organization
- [ ] Update command documentation if needed
- [ ] Update ./README.md documentation if needed

### Notes
- Package contains CLI command definitions and handlers
- Uses Cobra framework
- [ ] Specify additional cmd-specific tasks here

---

## Execution Order

**Recommended Order:**
1. Start with low-dependency packages: `internal/errors`, `internal/logger`, `internal/version`
2. Move to mid-level packages: `internal/config`, `internal/rclone`
3. Finish with high-level packages: `internal/sync`, `internal/cmd`

**Reasoning:**
- Low-dependency packages are easier to refactor
- Other packages depend on them, so they benefit from a stable base
- High-level packages may need to be updated after lower-level refactoring

---

## Completion Criteria

For each package:
- [ ] All common refactoring tasks completed
- [ ] All package-specific tasks completed
- [ ] All tests pass
- [ ] Logger consolidated (if applicable)
- [ ] Specification verified (if applicable)
- [ ] Code is simplified and follows DRY principle

---

## Notes for User

This is a **template plan** - review and customize:

- Add package-specific tasks to each milestone
- Adjust execution order if needed
- Add additional constraints or requirements
- Remove milestones for packages that don't need refactoring
- Update completion criteria based on your priorities
