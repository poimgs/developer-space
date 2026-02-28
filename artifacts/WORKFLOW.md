# Project Workflow

How this project's specs were created, step by step.

## Step 1: High-Level Overview (Gemini)

Start with a conversational brainstorm in Gemini. Describe the project idea — what it does, who it's for, rough feature set. Gemini helps shape the vision into a structured document: project overview, feature list, user roles, and tech preferences.

**Output:** A prompt summarizing the project scope and decisions, formatted for passing into Claude Code.

## Step 2: Technical Specification (Claude Code)

Pass the Gemini-generated prompt into Claude Code. Discuss technical details interactively — data model, API design, auth flow, deployment strategy, notification approach. Make decisions on ambiguous points (admin model, email provider, notification channel, spec format).

**Output:** The `specs/` directory — 8 feature-based markdown files covering the full application spec with user stories, acceptance criteria, API contracts, and implementation notes.

## Summary

```
Gemini                          Claude Code
  │                                │
  │  Brainstorm project idea       │
  │  Shape high-level overview     │
  │                                │
  │  Output: structured prompt ───▶│
  │                                │  Discuss technical decisions
  │                                │  Define data model & APIs
  │                                │  Write feature specs
  │                                │
  │                                │  Output: specs/ directory
```

| Step | Tool | Input | Output |
|------|------|-------|--------|
| 1 | Gemini | Project idea + conversation | Structured prompt for Claude Code |
| 2 | Claude Code | Gemini prompt + interactive discussion | `specs/00-overview.md` through `specs/07-deployment.md` |
