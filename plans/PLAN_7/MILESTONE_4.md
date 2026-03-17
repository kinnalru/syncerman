---
title: "Milestone 4: Refactor internal/config"
status: "✅ Completed"
---

# Milestone 4: Refactor internal/config

## Goal

Remove code duplication, simplify order-preserving logic, consolidate conversion functions, improve error messages, and verify test coverage for the internal/config package.

## Context

This milestone is part of PLAN_7: Refactoring Internal Packages. The config package handles configuration loading, parsing, and validation.

**Package State Analysis:**
- 7 files: doc.go, types.go, loader.go, validator.go, discovery.go, config_test.go, types_order_test.go (2,171 lines)
- Core implementation: 1,089 lines of production code
- Tests: 1,082 lines of test code
- Test coverage: Currently very high (> 97%)
- No local logger implementation (correctly uses internal/logger)
- No specification file exists

**Key Findings:**
- Order preservation critical for linear sync chains (local → gd → yd per OVERALL.md:364-403)
- Discovery functions duplicated between internal/config and internal/cmd
- Custom UnmarshalYAML implementations ensure YAML order preservation
- Configuration format: guides/OVERALL.md:46-111
- Code style guidelines: guides/STYLE.md

**Critical Issues Identified:**
1. ** duplications**: DiscoverConfigPath duplicated in internal/config and internal/cmd
2. **Order preservation**: Similar UnmarshalYAML implementations might consolidate
3. **Error messages**: Some lack context and helpful suggestions

**Documentation References:**
- guides/OVERALL.md:46-111 - Configuration format and schema
- guides/OVERALL.md:364-403 - Order preservation requirements
- guides/STYLE.md - Go coding style guidelines

## Tasks

### 1. Verify Logger Consolidation ✅

Completed: Verified that internal/config package has no logger implementations or imports. Package is pure configuration library that parses YAML, validates structures, and returns errors. Uses fmt.* only for error formatting. This is correct separation of concerns. No changes needed.

### 2. Remove Unused Code and Resolve Duplications ✅

Completed: Removed duplicate DiscoverConfigPath functions - kept robust version in internal/config/discovery.go, removed from internal/cmd/root.go. Removed deprecated GetProvidersMap() and related test. Removed 65 lines total. Updated internal/cmd/config.go to use robust version. All tests pass, coverage 97.0%.

### 3. Simplify Order-Preserving Logic ✅

Completed: Created unmarshalOrderedMap helper function to extract common logic from OrderedPaths and OrderedProviders UnmarshalYAML implementations. Reduced code by ~22 lines while preserving all functionality. All 52 tests pass, including order preservation tests. Linter clean.

### 4. Consolidate Conversion Functions ✅

Completed: Analyzed toPathMap (map[string][]Destination) and toOrderedPaths ([]Destination with order). Determined they are appropriate separate inverse operations with fundamental differences (order preservation). No actual logic duplication found. No consolidation needed. Follows DRY principle. All tests pass.

### 5. Review YAML Parsing Logic ✅

Completed: parseProviders is minimal wrapper (7 lines) that directly calls yaml.Unmarshal. Error handling is appropriate: parse returns raw errors, LoadConfig wraps with file/data path context, UnmarshalYAML adds type context, Validate adds detailed context. All edge cases handled. No changes needed. Implementation optimally separates concerns.

### 6. Improve Error Messages ✅

Completed: Improved all 16 error messages across the package. Added context, location information (file paths, line numbers), and actionable suggestions. Enhanced loader, validator, types, and discovery errors with examples and references. 71 tests pass, coverage 97.0%, linter clean.

### 7. Verify Test Coverage ✅

Completed: Verified comprehensive test coverage. 73 tests pass, coverage 97.8% (exceeds > 97% requirement). Edge cases, error validation, and order preservation thoroughly tested. Added 2 new tests (invalid YAML file, permission denied). 97.8% coverage achieved with only 3 small system-level error blocks uncovered (difficult to test).


Ensure all functionality is comprehensively tested. Check that test coverage remains high (> 97%) after refactoring. Add any missing tests for new or modified code.

- Run full test suite for config package
- Check coverage percentage
- Identify any uncovered lines
- Add tests for uncovered code
- Verify all edge cases are covered
