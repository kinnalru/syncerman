# Milestone 1: Preserve Configuration Target Execution Order

## Goal
Fix critical issue where targets are executed in random order instead of preserving configuration file order, enabling linear synchronization chains (A→B→C→D) to work correctly.

## Context

### Problem Statement
During SCENARIO1 testing, targets executed in non-deterministic order instead of configuration order, breaking linear synchronization:

**Configuration:**
```yaml
local:   # Should execute 1st
  '/path/to/local':
    - to: 'gd:syncerman/scenario1/'

gd:       # Should execute 2nd
  'syncerman/scenario1/':
    - to: 'yd:syncerman/scenario1/'

yd:       # Should execute 3rd
  'syncerman/scenario1/':
    - to: 'path/to/local2'
```

**Actual Execution (First Sync):**
1. gd→yd (when gd was empty - WRONG ORDER)
2. yd→local2 (when yd was empty - WRONG ORDER)
3. local→gd (executed correctly, but too late)

**Impact:**
- Files never propagate through entire chain
- yd and local2 remain empty
- Second sync fails with state corruption error
- Linear synchronization completely broken

### Root Cause Analysis (from REPORT.md)
- Targets likely stored in unordered map/hash
- YAML unmarshaling loses order
- Provider iteration not preserving configuration order
- Order is inconsistent between different executions

### Approach Selection
**Selected:** Approach from PLAN_5.md - Ordered Configuration Storage

**Reasoning:**
- Simple, focused fix
- No configuration format changes required
- Predictable results
- Minimal code changes
- Low risk

### Success Criteria from PLAN_5.md
1. ✅ Target execution order preserves configuration file order
2. ✅ Linear synchronization chains (A→B→C→D) work correctly
3. ✅ Data propagates through entire chain in single sync execution
4. ✅ All 4/8 SCENARIO1 criteria pass after fix
5. ✅ Second sync completes successfully
6. ✅ No state file corruption on subsequent runs
7. ✅ Backwards compatible with existing configurations
8. ✅ All existing tests continue to pass

## Tasks

### Task 1: Investigate Current Implementation
**Priority:** CRITICAL **Estimate:** 2h

**Objective:** Understand how targets are stored and iterated to identify where order is lost.

**Subtasks:**
1. Read internal/config/types.go to understand Config structure
2. Read internal/config/loader.go to see YAML unmarshaling
3. Read internal/sync/targets.go to see target expansion logic
4. Read internal/sync/execution.go to see main sync loop
5. Identify where provider order is determined
6. Identify where path order is determined
7. Identify where destination order is determined
8. Write findings to investigation notes

**Expected Outputs:**
- List of files/functions handling target order
- Identification of where order is lost (YAML unmarshal, map iteration, etc.)
- Assessment of current data structures
- Recommendation for ordered data structure approach

**Verification Strategy:**
- Confirm understanding of current implementation
- Identify exact point of order loss
- Confirm approach feasibility

**Context:**
- Configuration schema: (guides/OVERALL.md:46-99)
- Test scenario: (tests/SCENARIO1.md:1-359)
- Issue details: (REPORT.md:167-242)

---

### Task 2: Research Ordered Data Structure Solutions
**Priority:** HIGH **Estimate:** 1h

**Objective:** Research and evaluate Go options for ordered data structures that preserve YAML configuration order.

**Subtasks:**
1. Research yaml.v3 library ordering capabilities
2. Evaluate using slices instead of maps for providers
3. Evaluate using ordered map implementations
4. Research third-party ordered map libraries
5. Check go-yaml library for ordered unmarshaling
6. Evaluate complexity of each approach
7. Document pros/cons for each option

**Expected Outputs:**
- Comparison table of ordered data structure options
- Recommendation for best approach
- Example code snippets for chosen approach
- Risk assessment for each option

**Verification Strategy:**
- Confirm YAML library supports ordered unmarshaling
- Verify chosen approach is feasible in Go
- Ensure no breaking changes to config format

**Options to Evaluate:**

| Option | Description | Pros | Cons | Complexity |
|--------|-------------|-------|-------|------------|
| yaml.v3 with Slice | Use slices for providers | Preserves order | Must handle duplicates differently | LOW |
| OrderedMap lib | Third-party library | Ready solution | External dependency | MEDIUM |
| Custom struct | Define Config with ordered fields | No external deps | Breaking change to config format | HIGH |
| Sequence tracking | Track order separately | Minimal changes | Extra bookkeeping | MEDIUM |

**Expected Decision:** Choose yaml.v3 with Slice approach (ordered by YAML order)

**Context:**
- Code style guidelines: (guides/STYLE.md)
- Best practices for data structures

---

### Task 3: Update Configuration Type Definitions
**Priority:** CRITICAL **Estimate:** 2h

**Objective:** Modify internal/config/types.go to use ordered data structures that preserve YAML order.

**Subtasks:**
1. Add ordered ProviderSlice type (if using slice approach)
2. Replace ProviderMap with ProviderSlice in Config struct
3. Update GetProviders() to return ordered list
4. Update GetPaths(provider) to return ordered list
5. Update GetDestinations(provider, path) to return ordered list
6. Ensure backward compatibility with existing code
7. Add type-level comments documenting ordering preservation
8. Run go fmt on modified file

**Code Changes:**
```go
// Before (unordered):
type Config struct {
    Providers map[string]Provider `yaml:",inline"`
}

// After (ordered):
type Config struct {
    Providers []ProviderWithKey `yaml:",inline"`  // Preserve YAML order
}

type ProviderWithKey struct {
    Name string   `yaml:"-"`
    Data Provider `yaml:",inline"`
}
```

**Expected Outputs:**
- Modified internal/config/types.go
- Updated type definitions with ordered structures
- Backward-compatible accessor methods
- Formatted code following STYLE.md

**Verification Strategy:**
- Compile with no errors
- Run existing config tests
- Confirm ordered behavior in test context

**Context:**
- Current types: (internal/config/types.go:1-250)
- YAML format: (guides/OVERALL.md:46-88)
- Style guidelines: (guides/STYLE.md:26-113)

---

### Task 4: Update Configuration Loader
**Priority:** CRITICAL **Estimate:** 2h

**Objective:** Modify internal/config/loader.go to unmarshal YAML into ordered data structures.

**Subtasks:**
1. Update LoadConfig() to parse YAML preserving order
2. Handle provider name extraction from YAML keys
3. Maintain ProviderWithKey structure
4. Update parseProviderMap() if needed
5. Add error handling for order-related issues
6. Update LoadConfigFromData() similarly
7. Add logging for order preservation (debug level)
8. Run go fmt and goimports

**Code Changes:**
```go
// Load YAML with ordered unmarshaling
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, NewConfigError("failed to read config file", err)
    }
    
    var orderedConfig OrderedConfig
    if err := yaml.Unmarshal(data, &orderedConfig); err != nil {
        return nil, NewConfigError("invalid YAML format", err)
    }
    
    return orderedConfig.ToConfig(), nil
}
```

**Expected Outputs:**
- Modified internal/config/loader.go
- YAML unmarshaling preserving order
- Provider names properly extracted
- All existing config loading tests pass

**Verification Strategy:**
- Test loading existing config files
- Verify order preservation in loaded config
- Run tests in internal/config/*_test.go

**Context:**
- Current loader: (internal/config/loader.go:1-110)
- Loading tests: (internal/config/loader_test.go:1-100)

---

### Task 5: Update Sync Target Expansion
**Priority:** HIGH **Estimate:** 1h

**Objective:** Modify internal/sync/targets.go to work with ordered provider/path structures.

**Subtasks:**
1. Update ValidateTargets() to handle ordered providers
2. Update ExpandTargets() to preserve order
3. Ensure provider/path combinations iterate in order
4. Update FormatRemote() if needed
5. Maintain existing error handling
6. Add inline comments explaining order preservation
7. Run go fmt

**Code Changes Around:**
```go
// Ordered iteration over targets
func ExpandTargets(cfg *config.Config, targetFilter string) ([]Target, error) {
    var targets []Target
    
    for _, provider := range cfg.GetProviders() {  // GetProviders() returns ordered list
        providerName := provider.Name
        
        for _, pathItem := range provider.Data.GetPaths() {  // GetPaths() returns ordered list
            // ... expand destinations, preserve order
        }
    }
    
    return targets, nil
}
```

**Expected Outputs:**
- Modified internal/sync/targets.go
- Ordered target expansion logic
- All existing target tests pass

**Verification Strategy:**
- Unit test target expansion with ordered config
- Verify output list matches configuration order
- Run internal/sync/targets_test.go

**Context:**
- Current targets: (internal/sync/targets.go:1-210)
- Target tests: (internal/sync/targets_test.go:1-200)

---

### Task 6: Update Sync Execution Loop
**Priority:** CRITICAL **Estimate:** 1-2h

**Objective:** Modify internal/sync/execution.go to iterate targets in returned order (already preserved).

**Subtasks:**
1. Verify RunAll() iterates targets in provided order (should already work)
2. Update logging to show execution order
3. Add debug output showing target order
4. Update error handling if needed
5. Ensure sequential execution is preserved
6. Add documentation about order preservation
7. Run go fmt

**Note:** If Tasks 3-5 preserve order correctly in data structures, this task may be minimal.

**Expected Outputs:**
- Modified internal/sync/execution.go
- Ordered target iteration
- Enhanced logging for execution order
- All existing sync tests pass

**Verification Strategy:**
- Unit test sync execution order
- Verify logging shows correct order
- Run internal/sync/execution_test.go

**Context:**
- Current execution: (internal/sync/execution.go:1-120)
- Execution tests: (internal/sync/execution_test.go:1-150)

---

### Task 7: Add Unit Tests for Order Preservation
**Priority:** MEDIUM **Estimate:** 2h

**Objective:** Add comprehensive unit tests to verify target execution order matches configuration order.

**Subtasks:**
1. Test Config.GetProviders() returns ordered list
2. Test Config.GetPaths() returns ordered list 3. Test Config.GetDestinations() returns ordered list
4. Test target expansion preserves order
5. Test sync execution preserves order
6. Test linear sync chain configuration (2 hops)
7. Test linear sync chain configuration (3 hops)
8. Add table-driven tests for various order cases

**Test File:** internal/config/types_order_test.go (new)

**Example Tests:**
```go
func TestConfigProvidersOrderByYAML(t *testing.T) {
    yamlData := `
gd:
  "path1":
    - to: "remote:path1"
local:
  "path2":
    - to: "remote:path2"
    `
    // Verify GetProviders() returns [gd, local] in YAML order
}

func TestLinearThreeHopChain(t *testing.T) {
    config := createThreeHopConfig()
    targets := ExpandTargets(config, "")
    // Verify targets: local→A, A→B, B→C (chain order)
}
```

**Expected Outputs:**
- New test file internal/config/types_order_test.go
- Comprehensive order preservation tests
- All tests pass

**Verification Strategy:**
- Run new unit tests
- Ensure 100% test pass rate
- Check test coverage for order-related code

**Context:**
- Testing guidelines: (guides/STYLE.md:173-183)
- Existing test patterns: (internal/config/*_test.go)

---

### Task 8: Run Full Test Suite and Verify
**Priority:** CRITICAL **Estimate:** 2h

**Objective:** Execute entire test suite to ensure no regressions and verify fix works.

**Subtasks:**
1. Run all config tests: `go test ./internal/config/...`
2. Run all sync tests: `go test ./internal/sync/...`
3. Run all cmd tests: `go test ./internal/cmd/...`
4. Fix any test failures
5. Run linting: `make lint`
6. Run go vet: `go vet ./...`
7. Update test coverage
8. Document any test changes

**Expected Outputs:**
- All existing tests pass
- New order tests pass
- No lint errors
- Zero critical issues

**Verification Strategy:**
- 100% test pass rate
- Coverage maintained or improved
- Linter passes

**Context:**
- Build commands: (AGENTS.md:22-32)
- Linting: (AGENTS.md:50-66)
- Test execution: (AGENTS.md:34-48)

---

### Task 9: Execute SCENARIO1 and Verify Fix
**Priority:** CRITICAL **Estimate:** 1h

**Objective:** Run full SCENARIO1 test to verify the fix resolves all issues.

**Subtasks:**
1. Check that test environment is clean (or clean it)
2. Run full SCENARIO1 script from tests/SCENARIO1.md
3. Verify Step 1 passes (configuration check)
4. Verify Step 3 passes (dry run)
5. Verify Step 4 executes (actual sync)
6. Verify Step 5 shows correct file propagation:   - local: 5 files
   - gd: 5 files   - yd: 5 files
   - local2: 5 files 7. Verify Step 6 passes (second sync completes)
8. Verify all 8 success criteria from SCENARIO1 pass
9. Document results

**Expected Results:**
```bash
# Verification of correctness:
echo "Files in each location should be identical:"
# local: 5 files
# gd:syncerman/scenario1/: 5 files
# yd:syncerman/scenario1/: 5 files  # FIXED - was empty before
# local2: 5 files                           # FIXED - was empty before
```

**Exit Codes:** All steps should return 0

**Verification Strategy:**
- Manually verify all 8 criteria from SCENARIO1
- Confirm file contents match across all locations
- Second sync completes successfully

**Context:**
- SCENARIO1 spec: (tests/SCENARIO1.md)
- Success criteria: (tests/SCENARIO1.md:351-359)
- Environment info: (REPORT.md:365-368)

---

### Task 10: Update Documentation
**Priority:** LOW **Estimate:** 1h

**Objective:** Update relevant documentation to reflect the fix and any behavior changes.

**Subtasks:**
1. Update AGENTS.md if any build/test commands changed
2. Update guides/OVERALL.md if needed (unlikely unless schema changed)
3. Add notes about order preservation in sync behavior section
4. Update any relevant code comments
5. Verify README.md doesn't need updates
6. Ensure all examples in documentation still correct

**Expected Outputs:**
- Updated documentation files (if any needed)
- Clear notes about order preservation
- All examples still work

**Verification Strategy:**
- Read through updated docs
- Ensure accuracy and clarity
- Confirm examples are still valid

**Context:**
- Documentation: (AGENTS.md, guides/*.md, README.md)

---

## Testing Strategy

### Pre-Implementation Testing
- None (investigation)

### Implementation Testing
- Unit tests for each task
- Continuous testing as changes are made
- Run smaller test suites frequently

### Post-Implementation Testing
- Full test suite execution
- SCENARIO1 integration test
- Manual verification of behavior
- Linting and formatting checks

### Regression Testing
- Run all existing test suites
- Verify no breaking changes to existing configs
- Test with various configuration patterns
- Test edge cases

---

## Acceptance Criteria

### Functional Requirements
1. ✅ Target execution matches configuration order 100%
2. ✅ Linear synchronization A→B→C→D works in one execution
3. ✅ Data propagates through entire chain
4. ✅ SCENARIO1 passes all 8 criteria
5. ✅ No state file corruption on re-run

### Quality Requirements
1. ✅ All existing tests pass (0 regressions)
2. ✅ New unit tests have ≥80% coverage for order logic
3. ✅ No lint or vet errors
4. ✅ Code follows STYLE.md guidelines

### Integration Requirements
1. ✅ Backward compatible with existing configs
2. ✅ No API breaking changes
3. ✅ All example configurations still work

---

## Risk Mitigation

### Risk: Incompatible Config Changes
**Probability:** MEDIUM
**Mitigation:**- Careful design in Task 3-4- Maintain existing accessor methods
- Extensive test suite (Task 8)

### Risk: YAML Library Limitations
**Probability:** LOW
**Mitigation:**- Research in Task 2- Have fallback approach ready
- Use well-established yaml.v3 library

### Risk: Performance Degradation
**Probability:** LOW
**Mitigation:**- Use efficient ordered structures- Only optimize if needed
- Benchmark if performance concerns arise

### Risk: Breaking Existing Tests
**Probability:** MEDIUM
**Mitigation:**- Run full test suite in Task 8- Fix issues as found
- Maintain backward compatibility

---

## Success Metrics

### Code Quality
- Zero lint errors
- Zero vet errors
- ≥80% test coverage for new code
- All code follows STYLE.md

### Functional Correctness
- SCENARIO1: 8/8 criteria pass (currently 4/8)
- Linear sync chains work correctly
- State file corruption resolved

### Performance
- No measurable performance degradation
- Sync execution time unchanged

---

## References

### Documentation References
- (PLAN_5.md) - Overall plan and approach selection
- (REPORT.md:167-242) - Issue #1 details and evidence
- (REPORT.md:206-222) - Issue #2 state file corruption details
- (guides/STYLE.md) - Code style guidelines
- (guides/OVERALL.md:46-99) - Configuration format reference
- (guides/PLANING.md) - Workflow guidelines

### Code References (to be investigated)
- internal/config/types.go - Configuration structure definitions
- internal/config/loader.go - YAML loading logic
- internal/sync/targets.go - Target expansion logic
- internal/sync/execution.go - Main sync execution loop
- internal/cmd/sync.go - CLI sync command

### Test References
- tests/SCENARIO1.md - Full test scenario specification
- internal/config/*_test.go - Config unit tests
- internal/sync/*_test.go - Sync unit tests
- internal/cmd/root_test.go - CLI integration tests

---

## Notes

### Implementation Assumptions
1. yaml.v3 library supports ordered unmarshaling (to be confirmed in Task 2)
2. Provider and path order is both important
3. Backward compatibility requires maintaining existing accessor methods
4. No configuration file format changes desired

### Known Limitations
1. Fixing target order doesn't address validation warnings (separate issue)
2. State file cleanup may still be needed for corrupted scenarios
3. Complex multi-chain scenarios not tested

### Future Enhancements (Out of Scope)
1. Automatic dependency detection and ordering
2. Validation warnings for linear sync chains
3. State file recovery tools
4. Visualization of sync chains

---

**Milestone Status:** PENDING - Task 1 starts implementation

**Last Updated:** 2026-03-15
