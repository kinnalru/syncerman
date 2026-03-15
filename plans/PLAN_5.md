# PLAN_5: Fix Linear Synchronization Target Execution Order

## Overview
This plan addresses critical issues discovered during SCENARIO1 testing where targets are executed in non-deterministic order instead of configuration order, breaking linear synchronization chains.

## Root Problem
Targets are processed in random order instead of preserving configuration file order, making linear synchronization (A→B→C→D) impossible because data cannot propagate through the chain.

## Test Evidence

### SCENARIO1 Results (2026-03-15)
**Configuration:** Sequential chain local → gd → yd → local2

**First Sync Execution Order (WRONG):**
1. gd:syncerman/scenario1/ → yd:syncerman/scenario1/ (gd was empty, nothing synced)
2. yd:syncerman/scenario1/ → /path/to/local2 (yd was empty, nothing synced)
3. local:/path/to/local → gd:syncerman/scenario1/ (executed correctly, but too late)

**Second Sync Execution:**
1. local:/path/to/local → gd:syncerman/scenario1/ (SUCCESS - no changes needed)
2. gd:syncerman/scenario1/ → yd:syncerman/scenario1/ (FAILED - state file corruption)
3. Sync stopped due to previous target failure

### Impact
- Files only synced to gd (first link in chain)
- yd and local2 remained empty
- Linear synchronization chain completely broken
- Second sync fails permanently due to state file corruption
- **Critical blocker:** Non-deterministic order makes A→B→C→D scenarios unusable

## Issues Identified

### Issue #1: Non-Deterministic Target Execution Order
**Priority:** CRITICAL
**Type:** Bug
**Impact:** COMPLETE - Linear synchronization unworkable

**Symptoms:**
- Targets execute in different order than specified in YAML
- Order is inconsistent between different executions
- Suggests use of unordered data structures (maps/hashes)

**Test Evidence:**
- Execution 1: gd→yd, yd→local2, local→gd
- Execution 2: local→gd, gd→yd (different order)
- Both orders wrong for linear sync use case

**Configuration Used:**
```yaml
local:
  'path/to/local':
    - to: 'gd:syncerman/scenario1/'       # Should be 1st

gd:
  'syncerman/scenario1/':
    - to: 'yd:syncerman/scenario1/'          # Should be 2nd

yd:
  'syncerman/scenario1/':
    - to: 'path/to/local2'               # Should be 3rd
```

### Issue #2: State File Corruption Handling
**Priority:** HIGH
**Type:** Error Handling
**Impact:** SECOND - Subsequent syncs fail permanently

**Symptoms:**
- After first sync with wrong order, second sync fails on gd→yd
- Error: "Empty prior Path1 listing. Cannot sync to an empty directory"
- State file shows gd has files but yd is empty
- Mismatch cannot be resolved without manual cleanup

**Test Evidence:**
```bash
ERROR: Empty prior Path1 listing
ERROR: Cannot sync to an empty directory: /home/llm/.cache/rclone/bisync/gd_syncerman_scenario1..yd_syncerman_scenario1.path1.lst
ERROR: Bisync critical error: empty prior Path1 listing
[ERROR] Result: Command failed with exit code 7
```

### Issue #3: Missing Linear Sync Validation
**Priority:** MEDIUM
**Type:** Feature Gap
**Impact:** MINOR - Not blocking but user-hostile

**Problem:** No detection or warning when linear sync patterns are configured but execution order may break them.

## Success Criteria

### Milestone Success Criteria
1. ✅ Target execution order preserves configuration file order
2. ✅ Linear synchronization chains (A→B→C→D) work correctly
3. ✅ Data propagates through entire chain in single sync execution
4. ✅ All 4/8 SCENARIO1 criteria pass after fix
5. ✅ Second sync completes successfully
6. ✅ No state file corruption on subsequent runs
7. ✅ Backwards compatible with existing configurations
8. ✅ All existing tests continue to pass

## Non-Goals (Out of Scope)

- Adding new features beyond fixing the order issue
- Performance optimization (unless required for fix)
- Complete refactor of sync engine
- Changes to configuration file format
- Adding automatic dependency resolution (future enhancement)

## Approaches Considered

### Approach A: Ordered Configuration Storage (RECOMMENDED)
**Description:** Replace unordered maps with ordered data structures that preserve insertion order.

**Pros:**
- Simple, focused fix
- No configuration format changes
- Predictable results
- Easy to test and verify
- Minimal code changes

**Cons:**
- Requires careful iteration logic
- May need custom ordered type for Go maps

**Complexity:** LOW
**Risk:** LOW

### Approach B: Topological Sort for Dependency Detection
**Description:** Detect sync dependencies (A→B, B→C) and automatically order by dependency graph.

**Pros:**
- Robust for complex sync networks
- Handles multiple parallel chains
- Future-proof for advanced scenarios
- Can detect circular dependencies

**Cons:**
- Complex to implement correctly
- Over-engineering for current requirement
- May introduce new bugs
- Longer development time

**Complexity:** MEDIUM
**Risk:** MEDIUM

### Approach C: Configuration Order Enforcement Option
**Description:** Add flag to force strict configuration order execution.

**Pros:**
- Optional backwards-compatible change
- Users can choose behavior
- Easy to test with flag

**Cons:**
- Adds configuration complexity
- Not fixing root cause
- Feature creep
- Default behavior still broken

**Complexity:** LOW
**Risk:** LOW

**Selected Approach:** Approach A (Ordered Configuration Storage)

## Implementation Strategy

1. Understand current target storage and iteration
2. Identify where order is lost (likely YAML unmarshaling or map iteration)
3. Implement ordered data structures
4. Update iteration logic to preserve order
5. Add tests for linear sync scenarios
6. Run SCENARIO1 to verify fix

## Testing Strategy

### Test Cases to Add
1. Test linear sync with 2 hops (A→B→C) - basic case
2. Test linear sync with 3 hops (A→B→C→D) - current SCENARIO1
3. Test target order preservation across multiple configurations
4. Test backwards compatibility with existing sync configs
5. Test state file handling on re-run
6. Test mixed scenarios (linear + independent targets)

### Test Re-Run
1. Execute entire SCENARIO1 with fix applied
2. Verify 8/8 success criteria pass
3. Verify SCENARIO1 passes multiple times (idempotency)

## Risks and Mitigations

### Risk 1: Breaking Existing Functionality
**Probability:** MEDIUM
**Impact:** HIGH

**Mitigation:**
- Run full test suite after changes
- Add regression tests
- Support both ordered and unordered modes
- Document behavior change

### Risk 2: Performance Degradation
**Probability:** LOW
**Impact:** LOW

**Mitigation:**
- Benchmark before/after
- Use efficient ordered data structures
- Optimize only if needed

### Risk 3: YAML Library Limitations
**Probability:** MEDIUM
**Impact:** HIGH

**Mitigation:**
- Research YAML library capabilities
- Have fallback approach ready
- May need custom unmarshaling

## Timeline

**Estimate:** Based on complexity analysis and test requirements

| Phase | Estimated Time | Notes |
|--------|-----------------|-------|
| Investigation | 2h | Understand current implementation |
| Design Decision | 1h | Confirm approach |
| Implementation | 4-6h | Ordered data structures + iteration |
| Testing | 2-3h | New tests + SCENARIO1 verification |
| Documentation | 1h | Update guides if needed |
| Total | 10-13h | ~1.5-2 days |

## References

### Related Documentation
- (guides/OVERALL.md:46-99) - Configuration format specification
- (guides/STYLE.md) - Go code style guidelines
- (guides/PLANING.md:1-34) - Workflow guidelines

### Related Files (to be investigated)
- internal/config/types.go - Configuration structure
- internal/config/loader.go - YAML loading logic
- internal/sync/targets.go - Target expansion
- internal/sync/execution.go - Main sync loop
- internal/cmd/sync.go - CLI sync command

### Test Files
- tests/SCENARIO1.md - Test specification
- internal/sync/*_test.go - Existing sync tests
- scenario test files for new test cases

## Next Steps

See MILESTONE1.md for detailed implementation plan.
