# Skill Authoring Guide

A skill is a checklist for one action or workflow step — not a reference manual. Skills complement `AGENTS.md`; they do not duplicate it. One skill = one behavior enforcement document.

## What a Skill Is (and Is Not)

**Is:** A focused, directive checklist for a single action (e.g., "create a UI component", "add a ctl-api function"). Under 80 lines. Answers: "When Claude is about to do X, what must Claude check or follow?"

**Is not:** An architecture reference, a duplicate of `AGENTS.md` content, or a general-purpose guide covering multiple actions.

## Naming Convention

Use action-oriented names, not agent-name mirrors:

- Good: `dashboard-ui-component`, `dashboard-ui-api`
- Bad: `dashboard-ui-builder` — mirrors an agent name, causes naming collision

Pattern for dashboard-ui skills: `dashboard-ui-<domain>`

Never name a skill identically to an existing agent.

## File Location

```
.agents/skills/<skill-name>/SKILL.md
```

Rules:
- Always a subdirectory under `.agents/skills/`, never a flat file
- Never nested deeper than one subdirectory
- No `rules/` subdirectory — no existing skill uses one and none are needed

## Required Front-Matter

```yaml
---
name: <skill-name>
description: Use this skill when <specific action>.
model: sonnet
color: blue
---
```

## Required Body Sections

1. **Scope anchor** — one sentence stating what the skill enforces
2. **Steps** — numbered, ordered, directive ("do X before Y" not "consider X or Y")
3. **Anti-Patterns** — what not to do and why

## Scope Discipline

- Keep skill body under 80 lines
- Do not repeat architecture content already in `AGENTS.md` — link to it instead
- Body must be directive, not exploratory
- If you find yourself writing more than 3 paragraphs of background context, stop — move it to `AGENTS.md`
- If a skill body is longer than the agent it relates to, it is too long

## Minimal Template (copy-paste starting template for new skills)

```markdown
---
name: <skill-name>
description: Use this skill when <specific action>.
model: sonnet
color: blue
---

This skill enforces <one-sentence scope>.

## Steps

1. <First mandatory step>
2. <Second mandatory step>
3. <Continue in order>

## Anti-Patterns

- **Do not** <bad thing> — <why it's bad>
- **Do not** <bad thing> — <why it's bad>
```

## Anti-Patterns for Skill Authors

- **Do not copy the full agent body into a skill.** See `.agents/skills/frontend-builder/SKILL.md` — this file predates the guide and is the canonical example of what to avoid. It is a 195-line verbatim copy of the agent body. New skills should be ~40–80 lines covering one workflow step.
- **Do not create a `rules/` subdirectory.** No existing skill uses one.
- **Do not write a skill that covers more than one distinct action.** If it feels like two things, make two skills.
- **Do not name a skill identically to an existing agent.**
- **Do not duplicate `AGENTS.md` content.** Link to it: "See AGENTS.md — TypeScript Conventions."
