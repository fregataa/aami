# 🤖 AI Coding Agent Documentation

This directory contains all documentation specifically designed for AI Coding Agents (Claude, GPT, etc.) working on the AAMI project.

**For general users and developers**, please refer to [docs/](../docs/) instead.

---

## 📁 Directory Structure

```
.agent/
├── README.md                           # This file - Agent documentation index
├── context/                            # Project context for agents
│   ├── project-overview.md             # High-level project description
│   ├── architecture.md                 # System architecture
│   └── coding-standards.md             # Coding conventions
├── planning/                           # Project planning documents
│   ├── README.md                       # Planning documentation index
│   ├── PLAN.md                         # Master project plan
│   ├── sprints/                        # Sprint-by-sprint plans
│   │   ├── sprint0-plan.md
│   │   ├── sprint1-plan.md
│   │   └── ... (sprint2-10)
│   └── tasks/                          # Task breakdown documents
│       └── sprint4-breakdown.md
└── services/                           # Service-specific agent documentation
    └── config-server/                  # Config Server agent docs
        ├── README.md                   # Config Server agent guide
        ├── refactoring/                # Refactoring documentation
        │   ├── target-group-relationship.md
        │   └── template-instance-decoupling.md
        └── sprint-plan.md              # Service-specific sprint plan
```

---

## 🎯 Purpose

This directory serves as the **central knowledge base for AI coding agents** working on AAMI. It contains:

1. **Context** - Understanding the project, architecture, and conventions
2. **Planning** - Project roadmap, sprint plans, and task breakdowns
3. **Service Guides** - Service-specific implementation details and refactoring plans

**Key Principle**: Agent documentation is kept separate from user-facing documentation to:
- Provide deeper technical context for autonomous coding work
- Track project planning and evolution
- Document refactoring decisions and architectural changes
- Maintain agent-specific guides without cluttering user docs

---

## 🚀 Getting Started (for Agents)

### First Time Working on This Project?

**Start with the root [AGENT.md](../AGENT.md)**, then read in this order:

1. **[context/project-overview.md](context/project-overview.md)** *(Coming soon)* - What is AAMI?
2. **[context/architecture.md](context/architecture.md)** *(Coming soon)* - How is it structured?
3. **[context/coding-standards.md](context/coding-standards.md)** *(Coming soon)* - What are the conventions?
4. **[planning/PLAN.md](planning/PLAN.md)** - What's the project roadmap?

### Working on a Specific Sprint?

Navigate to **[planning/sprints/](planning/sprints/)** and find the relevant sprint plan:
- Sprint 0-10 plans document objectives, tasks, and outcomes
- Task breakdowns in [planning/tasks/](planning/tasks/) provide detailed implementation guidance

### Working on Config Server?

Check **[services/config-server/](services/config-server/)**:
- **[README.md](services/config-server/README.md)** - Service architecture and coding patterns
- **[refactoring/](services/config-server/refactoring/)** - Ongoing refactoring documentation
- **[sprint-plan.md](services/config-server/sprint-plan.md)** - Service-specific sprint roadmap

---

## 📚 Document Types

### 1. Context Documents (`context/`)

**Purpose**: Provide foundational understanding of the project

**When to read**:
- First time working on the project
- When architectural decisions are unclear
- When coding patterns are inconsistent

**When to update**:
- Major architectural changes
- New coding conventions established
- Project scope changes

### 2. Planning Documents (`planning/`)

**Purpose**: Track project evolution and sprint progress

**When to read**:
- Starting work on a new feature
- Understanding project priorities
- Checking historical context for decisions

**When to update**:
- Sprint completion (add retrospective notes)
- Major plan changes
- New task breakdowns created

### 3. Service Documentation (`services/`)

**Purpose**: Service-specific implementation details and refactoring plans

**When to read**:
- Working on a specific service
- Understanding service architecture
- Planning service refactoring

**When to update**:
- Completing major refactoring
- Documenting important architectural decisions
- Adding new service-specific patterns

---

## 🔄 Workflow Guidelines

### When Starting a Task

1. **Read relevant context** - Understand the project and service architecture
2. **Check sprint plans** - See if your task is part of a sprint
3. **Review service docs** - Look for service-specific patterns
4. **Follow coding standards** - Maintain consistency

### When Completing Work

1. **Update sprint notes** - If task was part of a sprint
2. **Document decisions** - Add to service docs if significant
3. **Update context** - If you changed architecture or patterns
4. **Update AGENT.md** - If documentation structure changed

---

## ⚠️ Important Notes

### What NOT to Do

1. **Don't mix agent and user documentation**
   - Agent docs stay in `.agent/`
   - User docs stay in `docs/`

2. **Don't modify sprint plans retroactively**
   - Sprint plans are historical records
   - Add notes or retrospectives instead

3. **Don't skip context documents**
   - Understanding architecture prevents mistakes
   - Coding standards ensure consistency

4. **Don't document in comments what should be in docs**
   - Complex architectural decisions → context/ or services/
   - Implementation patterns → services/
   - User-facing guides → docs/

### Best Practices

1. **Always start with context** - Read before coding
2. **Follow existing patterns** - Check similar code
3. **Document significant decisions** - Help future agents
4. **Ask when uncertain** - Better to clarify than assume
5. **Keep documents up to date** - Stale docs are harmful

---

## 🔗 Related Documentation

- **Root**: [AGENT.md](../AGENT.md) - Quick orientation guide for agents
- **User Docs**: [docs/](../docs/) - User-facing documentation
- **Main README**: [README.md](../README.md) - Project overview for users

---

## 📝 Document Update Protocol

When you complete significant work, consider updating:

| Change Type | Update Location | Example |
|------------|----------------|---------|
| Architecture change | `context/architecture.md` | New microservice added |
| New pattern established | `context/coding-standards.md` | Error handling pattern |
| Sprint completion | `planning/sprints/sprint*.md` | Add retrospective notes |
| Major refactoring | `services/*/refactoring/*.md` | Document before/after |
| Service pattern change | `services/*/README.md` | New validation approach |

---

**Last Updated**: 2025-12-29
**Maintained By**: AI Coding Agents & Project Team

For questions or improvements to this documentation structure, discuss with the project team.
