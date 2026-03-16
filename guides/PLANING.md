# Autonomous Coding Agent Workflow

This document outlines the standard operating procedure for autonomous coding agents PLANNING to the project. Agents must follow these guidelines to ensure structured, safe, and verifiable progress. You MUST use `todowrite` and `todoread` tools.

## 1. Overall Planning with Multiple Milestones 

Always read `guides/OVERALL.md` when planning. You MUST use `todowrite` and `todoread` tools.

*   **Context Gathering**:
    - Begin by exploring the codebase, reading documentation, and analyzing the core request to understand the full scope of the work.
*   **Milestone Definition**:
    - Deconstruct the overarching goal into a sequence of logical, manageable milestones
    - Each milestone must represent a significant, standalone, and verifiable phase of the project (e.g., "Implement Data Layer", "Build Core API", "Integrate Frontend Components")
*   **Tasks**:
    - there must be 5-7 SHORT descriptive tasks in milestone when Overall Planning
    - try to avoid code snippets in `Plan` definition unless necessary
*   **Refferences**:
    - MAY Use refferences to key documents (guides/PLANING.md, and other) example: `(OVERALL.md: lines 256-267)`
*   **Dependency Mapping**:
    - Sequence milestones logically, ensuring that foundational work is completed before dependent features.
    - Planned milestones MUST be executed sequentaly
*   **Persistance**:
    - Persist `Plan` in Markdown file  `plans/PLAN_<NUM>.md` where NUM is sequental number
    - MUST use `Plan Format`
    - MUST use Markdown Frontmatter in `Plan` file (see `Plan Format`)
    - Allowed Plan states: `Pending`, `In Progress`, `Completed`, `Archived` 
    - MUST use icons: ⏸️, ⌛, ✅, 📦 for states
    - Maintain `Plan` status in Frontmatter during Plan/Milestone Execution
*   **Tracker Update**: 
    - Populate the active progress in `./WAL.md` (see Write Ahead Log definition an rules in `guides/WRITE_AHEAD_LOGS.md`) BEFORE and AFTER actual work
*   **Subagent invocation**:
    - MAY use tool with subagents → Use OpenCode's subagent system (@mention `@plan`). example: `@plan create a plan to implement Milestone 3`
    - MAY use the task tool with subagent_type set to "plan" and pass the task description.

### `Plan Format`

Persisted `plans/PLAN_<NUM>.md` use strict format from `guides/templates/PLAN_NUM.md`.

MUST USE `guides/templates/PLAN_NUM.md` template for `Plan` definition
MUST USE Markdown Frontmatter to track `Plan` status

## 2. Comprehensive Detailed Planning for a Milestone

Once a milestone is selected for execution, it must be broken down into concrete, actionable steps.
Always read `guides/OVERALL.md` when planning. You MUST use `todowrite` and `todoread` tools.

*   **Milestone Selection**:
    - Select next unfinished milestone sequentaly
*   **Deep Dive Analysis**:
    - Perform targeted exploration of the domain specific to the selected milestone
    - Understand existing file structures, interfaces, and patterns related to this phase.
    - MAY use external sources through websearch
*   **Task Decomposition**:
    - Break the milestone down into atomic tasks. An ideal task should be scoped to a single logical change (e.g., "Create database schema for User model", "Write unit tests for User schema").
    - try to avoid specific code snippets, use text description.
    - no more than 7 tasks in milestone
*   **Verification Strategy**:
    - For each milestone, clearly define how success will be verified. This should include specific testing, linting, or building steps that must pass upon completion.
*   **Refferences**:
    - MAY Use refferences to key documents (guides/PLANING.md, PLAN_NUM.md and other) lxample: `(PLAN_1.md: lines 11-13)`
*   **Persistance**:
    - Persist `Milestone` in Markdown file  `plans/PLAN_N/MILESTONE_<NUM>.md` where NUM is sequental number
    - MUST use `Milestone Format`
    - MUST use Markdown Frontmatter in `Milestone` file (see `Milestone Format`)
    - Allowed Milestone states: `Pending`, `In Progress`, `Completed`, `Archived` 
    - MUST use icons: ⏸️, ⌛, ✅, 📦 for states
    - Maintain `Milestone` status in Frontmatter during Milestone/Task Execution
*   **Tracker Update**:
    - Populate the active milestone in `./WAL.md` (see Write Ahead Log definition an rules in `guides/WRITE_AHEAD_LOGS.md`) with these granular tasks. Write BEFORE and AFTER actual work.
*   **Subagent invocation**:
    - MAY use tool with subagents → Use OpenCode's subagent system (@mention `@plan`). example: `@plan create a plan to implement Milestone 3`
    - MAY use the task tool FOR PLANING with subagent_type set to "plan" and pass the task description.
    - MAY use the task tool FOR EXECUTING with subagent_type set to "build" and pass the task description.

### `Milestone Format`

Persisted `plans/PLAN_N/MILESTONE_<NUM>.md` use strict format from `guides/templates/MILESTONE_NUM.md`.

MUST USE `guides/templates/MILESTONE_NUM.md` template for `Milestone` definition
MUST USE Markdown Frontmatter to track `Milstone` status

Allowed Task states: `Pending`, `In Progress`, `Completed`
MUST use icons: ⏸️, ⌛, ✅ for states

### `Task Format`

Use strict format from `guides/templates/TASK.md`.

MUST USE `guides/templates/TASK.md` template for `Task` definition

Task example:

```
### 5 Implement First-Run Error Detection and Handling

Add automatic first-run error handling:

- Integrate `IsFirstRunError()` from rclone package into sync engine
- Implement `HandleFirstRunError()` to retry with --resync flag
- Detect the "cannot find prior Path1 or Path2 listings" error pattern
- Log first-run detection to user
- Retry sync with --resync flag automatically
- Track if sync was retried for first-run
```


## 3. Executing Tasks from Milestones

Execution must be methodical and focused to maintain code quality. You MUST use `todowrite` and `todoread` tools.

*   **Single-Task Focus**:
    - Mark exactly *one* task as "In Progress" at a time
    - The agent must never work on multiple unrelated tasks simultaneously to prevent context switching and contamination.
*   **Implementation**:
    - Execute the necessary code changes adhering strictly to the project's style guides and conventions.
*   **Continuous Verification**:
    - Immediately after completing a task's code, run local verification steps (e.g., formatters, linters, unit tests).
*   **Task Completion**:
    - Once verified successfully, mark the task as "Completed" inside Milestone file.
*   **Dynamic Adaptation**:
    - If a task requires unforeseen subsequent steps, dynamically add new tasks to the current milestone rather than deviating from the plan.
*   **Subagent invocation**:
    - you MUST use subagents for each task wia `task` tool. 
    - MAY use tool with subagents → Use OpenCode's subagent system (@mention `@build`)
    - MAY use the task tool FOR EXECUTING with subagent_type set to "build" and pass the task description.

## 4. Fixing and Tracking Current Progress

Robust state management and error handling are critical for autonomous progress. Use `./WAL.md` as Write Ahead Log. see Write Ahead Log definition an rules in `guides/WRITE_AHEAD_LOGS.md`.

*   **Handling Failures**:
    - If a task fails verification (e.g., a test fails or the build breaks), the agent must not proceed to the next task.
      1.  Pause current progress on the original task.
      2.  Analyze the error output.
      3.  Create an immediate, high-priority sub-task to fix the issue.
      4.  Resolve the error and successfully re-verify before resuming the original plan.
*   **Checkpointing**:
    - Treat logical completion points (like finishing a significant task or a full milestone) as checkpoints. 
    - Ensure the codebase is in a stable, verifiable state before moving on. Fix checkpoint in Write Ahead Log.
*   **Single Source of Truth**:
    - The agent must treat its `./WAL.md` as the definitive record of progress.
    - The status of milestones and tasks must be continuously updated in real-time to reflect the actual state of the codebase.
*   **Milestone completition**:
    - Mark all tasks in `plans/PLAN_N/MILESTONE_<NUM>.md` after `Mielstone` (all its tasks) completition
*   **Plan completition**:
    - Mark all milestones in `plans/PLAN_<NUM>.md` after `Plan` (all its milestones) completition

## 5. Flow and Steps definition

**General Flow**:
1. List all `plans/**/*`
2. read Write Ahead Log and find next unfinished `Task` or `Milestone`
3. If there is unfinished `Task` or `Milestone` use `Continue Current Milestone Flow`
4. If there is no unfinished `Task` or `Milestone` use `Take Next Milestone Flow`
5. when all Milestones from `plans/PLAN_N/MILESTONE_<NUM>.md` finished use 'Create New Milestone Flow'


**Continue Current Milestone Flow**:
1. read current `Milestone` tasks from `plans/PLAN_N/MILESTONE_<NUM>.md`
2. execute all unfinished `Tasks` in sequental order
3. when all Tasks finished use `Take Next Milestone Flow`


**Take Next Milestone Flow**:
1. List all `plans/**/*`
2. read next UNFINISHED `Milestone` from `plans/PLAN_N/MILESTONE_<NUM>.md`
3. execute all unfinished Tasks in sequental order
4. when all Tasks finished use `Take Next Milestone Flow`
5. when all Milestones from `plans/PLAN_N/MILESTONE_<NUM>.md` finished use `Create New Milestone Flow`


**Create New Milestone Flow**:
1. CLEAN Write head Log:
   - remove details from completed milestones leave only names and status. example: `### 2026-03-14 Milestone 1: Project Foundation and Core Structure - COMPLETED`
2. List all `plans/**/*`
3. read next UNFINISHED `plans/PLAN_<NUM>.md`
4. created detailed complehensive Milestone definitions in `plans/PLAN_N/MILESTONE_<NUM>.md`
5. stop working when all milestones and plans completed
6. use `Take Next Milestone Flow` 

## 6. subagents spawning

**When planing:**
You MUST use tool with subagents → Use OpenCode's subagent system (@mention `@plan`)
I should use the task tool with subagent_type set to "plan" and pass the task description.

**When building/executing tasks:**
You MUST use tool with subagents → Use OpenCode's subagent system (@mention `@build`)
I should use the task tool with subagent_type set to "build" and pass the task description.


## 7. Persistant planing structure example:

`tree plans`
```
plans
├── PLAN_1
│   ├── MILESTONE_1.md
│   ├── MILESTONE_2.md
│   ├── MILESTONE_3.md
│   ├── MILESTONE_4.md
│   ├── MILESTONE_5.md
│   └── MILESTONE_6.md
├── PLAN_1.md
├── PLAN_2.md
└── PLAN_3.md
```
