# Autonomous Coding Agent Workflow

This document outlines the standard operating procedure for autonomous coding agents PLANNING to the project. Agents must follow these guidelines to ensure structured, safe, and verifiable progress.

## 1. Overall Planning with Multiple Milestones

Before writing any code, the agent must establish a high-level roadmap.
Always read `guides/OVERALL.md` when planning.

*   **Context Gathering:** Begin by exploring the codebase, reading documentation, and analyzing the core request to understand the full scope of the work.
*   **Milestone Definition:** Deconstruct the overarching goal into a sequence of logical, manageable milestones. Each milestone must represent a significant, standalone, and verifiable phase of the project (e.g., "Implement Data Layer", "Build Core API", "Integrate Frontend Components").
*   **Refferences:** MAY Use refferences to key documents (guides/PLANING.md, and other) lxample: `(OVERALL.md: lines 256-267)`
*   **Dependency Mapping:** Sequence milestones logically, ensuring that foundational work is completed before dependent features.
*   **State Tracking:** Maintain a structured list of milestones, tracking their current state (e.g., Pending, In Progress, Completed).
*   **Approval:** Present the high-level roadmap for user alignment and approval before proceeding to implementation.
*   **Storage** Persist in Markdown file  `plans/PLAN_<NUM>.md` where NUM is sequental number
*   **Tracker Update:** Populate the active progress in `./WAL.md` (Write Ahead Log) BEFORE and AFTER actual work


## 2. Comprehensive Detailed Planning for a Milestone

Once a milestone is selected for execution, it must be broken down into concrete, actionable steps.
Always read `guides/OVERALL.md` when planning.

*   **Milestone Selection:** Select the highest priority pending milestone that has its dependencies met.
*   **Deep Dive Analysis:** Perform targeted exploration of the domain specific to the selected milestone. Understand existing file structures, interfaces, and patterns related to this phase.
*   **Task Decomposition:** Break the milestone down into atomic tasks. An ideal task should be scoped to a single logical change (e.g., "Create database schema for User model", "Write unit tests for User schema").
*   **Verification Strategy:** For each milestone, clearly define how success will be verified. This should include specific testing, linting, or building steps that must pass upon completion.
*   **Refferences:** MAY Use refferences to key documents (guides/PLANING.md, PLAN_NUM.md and other) lxample: `(PLAN_1.md: lines 11-13)`
*   **Tracker Update:** Populate the active milestone in `./WAL.md` (Write Ahead Log) with these granular tasks. Write BEFORE and AFTER actual work.
*   **Storage:** Persist in Markdown file  `plans/MILESTONE_<NUM>.md` where NUM is milestone number.
*   **Subagent invocation:** use `@plan` in planing agent invocation. example: `@plan create a plan to implement Milestone 3`

Format:

```
# Milestone <MILESTONE NUM>: <MILESTONE NAME>

## Goal

<GOAL OF MILESTONE>

## Context

<CONTEXT OF MILESTONE (docs, references, ...)>

## Tasks

### <TASK NUM>. <TASK NAME>

<TASK DEFINITION>
```

## 3. Executing Tasks from Milestones

Execution must be methodical and focused to maintain code quality.

*   **Single-Task Focus:** Mark exactly *one* task as "In Progress" at a time. The agent must never work on multiple unrelated tasks simultaneously to prevent context switching and contamination.
*   **Implementation:** Execute the necessary code changes adhering strictly to the project's style guides and conventions.
*   **Continuous Verification:** Immediately after completing a task's code, run local verification steps (e.g., formatters, linters, unit tests).
*   **Task Completion:** Once verified successfully, mark the task as "Completed".
*   **Dynamic Adaptation:** If a task requires unforeseen subsequent steps, dynamically add new tasks to the current milestone rather than deviating from the plan.
*   **Subagents** you MUST use subagents for each task wia `task` tool. 

## 4. Fixing and Tracking Current Progress

Robust state management and error handling are critical for autonomous progress. Use `./WAL.md` as Write Ahead Log.

*   **Handling Failures:** If a task fails verification (e.g., a test fails or the build breaks), the agent must not proceed to the next task.
    1.  Pause current progress on the original task.
    2.  Analyze the error output.
    3.  Create an immediate, high-priority sub-task to fix the issue.
    4.  Resolve the error and successfully re-verify before resuming the original plan.
*   **Checkpointing:** Treat logical completion points (like finishing a significant task or a full milestone) as checkpoints. Ensure the codebase is in a stable, verifiable state before moving on. Fix checkpoint in Write Ahead Log.
*   **Single Source of Truth:** The agent must treat its `./WAL.md` as the definitive record of progress. The status of milestones and tasks must be continuously updated in real-time to reflect the actual state of the codebase.
*   **Milestone completition** Mark all tasks in `plans/MILESTONE_<NUM>.md` after completition. USE Frontmatter for marking
*   

## 5. Flow and Steps definition

**General Flow**:
1. read Write Ahead Log and find next unfinished Task or Milestone
2. If there is unfinished TASK or MILESTONE  use 'Continue Current Milestone Flow'
3. If there is no unfinished TASK or MILESTONE use 'Take Next Milestone Flow'
4. when all Milestones from `plans/MILESTONE_<NUM>.md` finished use 'Create New Milestone Flow'


**Continue Current Milestone Flow**:
1. read current Milestone tasks from `plans/MILESTONE_<NUM>.md`
2. execute all unfinished Tasks in sequental order
3. when all Tasks finished use 'Take Next Milestone Flow'


**Take Next Milestone Flow**:
1. read next UNFINISHED Milestone from `plans/MILESTONE_<NUM>.md`
2. execute all unfinished Tasks in sequental order
3. when all Tasks finished use 'Take Next Milestone Flow'
4. when all Milestones from `plans/MILESTONE_<NUM>.md` finished use 'Create New Milestone Flow'


**Create New Milestone Flow**:
1. CLEAN Write head Log - remove details from completed milestones leave only names and status. example: `### 2026-03-14 Milestone 1: Project Foundation and Core Structure - COMPLETED`
2. read next UNFINISHED `plans/PLAN_<NUM>.md`
3. created detailed complehensive Milestone definitions in `plans/MILESTONE_<NUM>.md`
4. use 'Take Next Milestone Flow'