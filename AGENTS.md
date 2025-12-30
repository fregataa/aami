# 🤖 AI Coding Agent Guide

> **For AI Coding Agents (Claude, GPT, etc.)**  
> This document helps you quickly understand the AAMI project structure and locate relevant documentation.

## Quick Orientation

**Project**: AAMI (Autonomous Agent Management Infrastructure)  
**Type**: Infrastructure monitoring and configuration management system  
**Main Language**: Go  
**Architecture**: Microservices with config-server as the core component

## 📁 Documentation Structure

### For AI Agents (You are here!)

All agent-specific documentation is in the **`.agent/`** directory:

```
.agent/
├── README.md                    # Agent documentation index
├── context/                     # Project context for agents
│   ├── project-overview.md      # High-level project description
│   ├── architecture.md          # System architecture
│   └── coding-standards.md      # Coding conventions
├── planning/                    # Project planning documents
│   ├── PLAN.md                  # Master project plan
│   ├── README.md                # Planning index
│   ├── sprints/                 # Sprint plans
│   │   ├── sprint0-plan.md
│   │   ├── sprint1-plan.md
│   │   └── ... (sprint0-10)
│   └── tasks/                   # Task breakdown documents
│       └── sprint4-breakdown.md
└── services/                    # Service-specific agent docs
    └── config-server/           # Config Server agent docs
        ├── README.md            # Config Server agent guide
        ├── refactoring/         # Refactoring documentation
        │   ├── target-group-relationship.md
        │   └── template-instance-decoupling.md
        └── sprint-plan.md       # Service sprint plan
```

### For Human Users

User-facing documentation is in the **`docs/`** directory:

```
docs/
├── README.md                    # Documentation index
├── en/                          # English documentation
│   ├── QUICKSTART.md
│   ├── DEVELOPMENT.md
│   ├── API.md
│   └── ...
├── kr/                          # Korean documentation
│   └── ...
└── diagrams/                    # Architecture diagrams
```

---

## 📝 Document Update Protocol

**⚠️ CRITICAL**: Read this section carefully before creating or modifying any agent documentation.

When you complete work:

1. **Context changes**: Update `.agent/context/`
2. **Sprint completion**: Add notes to `.agent/planning/sprints/`
3. **Refactoring**: Document in `.agent/services/*/refactoring/`
4. **New patterns**: Update `.agent/context/coding-standards.md`

### ⚙️ Creating Planning Documents

**IMPORTANT**: All planning documents must be created under `.agent/` directory to prevent git tracking.

When creating new planning documents:

- **Sprint Plans** → `.agent/planning/sprints/`
  - Example: `.agent/planning/sprints/sprint11-plan.md`

- **Task Breakdowns** → `.agent/planning/tasks/`
  - Example: `.agent/planning/tasks/sprint5-breakdown.md`

- **Refactoring Plans** → `.agent/services/{service-name}/refactoring/`
  - Example: `.agent/services/config-server/refactoring/api-redesign.md`

- **Implementation Plans** → `.agent/planning/` or service-specific directory
  - Example: `.agent/planning/migration-plan.md`

**Why?**
- Planning documents are for agent context, not end-users
- Prevents cluttering the main repository
- Keeps agent workspace organized
- `.agent/` directory should be in `.gitignore`

**Never create planning documents in:**
- ❌ Root directory
- ❌ `docs/` directory (reserved for user documentation)
- ❌ Service root directories (use `.agent/services/` instead)

### ✍️ Writing Guidelines for Agent Documents

**IMPORTANT**: Agent documents must be written for AI agents to understand easily.

When writing or updating agent documentation:

1. **Be Explicit and Structured**
   - Use clear headings and sections
   - Provide step-by-step instructions
   - Include code examples where applicable
   - Use bullet points and numbered lists

2. **Provide Context**
   - Explain the "why" behind decisions
   - Link to related documentation
   - Reference specific file paths with line numbers if relevant
   - Include before/after examples for refactoring

3. **Use Consistent Formatting**
   - Use markdown formatting consistently
   - Code blocks with language identifiers (```go, ```bash, etc.)
   - Tables for comparing options or listing specifications
   - Diagrams or ASCII art for architecture

4. **Be Precise**
   - Use exact file paths (e.g., `services/config-server/internal/errors/errors.go`)
   - Include function/struct names with line numbers (e.g., `FromGormError:45`)
   - Specify versions when mentioning dependencies
   - Provide concrete examples over abstract descriptions

5. **Make It Actionable**
   - Every document should have clear next steps
   - Include commands that can be copy-pasted
   - Provide checklists for multi-step tasks
   - List acceptance criteria for tasks

**Remember**: Agent documents are working documents for autonomous coding. They should enable an agent to understand context and take action without human intervention.

---

## 🎯 Getting Started (for Agents)

### 1. First Time Working on This Project?

Read in this order:
1. **This file** (AGENT.md) - You're here! ✓
2. **[.agent/README.md](.agent/README.md)** - Agent documentation index
3. **[.agent/context/project-overview.md](.agent/context/project-overview.md)** - Project overview
4. **[.agent/context/architecture.md](.agent/context/architecture.md)** - System architecture
5. **[.agent/planning/PLAN.md](.agent/planning/PLAN.md)** - Master plan

### 2. Working on a Specific Sprint?

Check: **[.agent/planning/sprints/](.agent/planning/sprints/)**
- Find the current sprint plan (sprint0-10)
- Review sprint objectives and tasks

### 3. Working on Config Server?

Check: **[.agent/services/config-server/](.agent/services/config-server/)**
- README.md for service-specific guidance
- refactoring/ for ongoing refactoring work

### 4. Need Coding Standards?

Check: **[.agent/context/coding-standards.md](.agent/context/coding-standards.md)**
- Go coding conventions
- Error handling patterns
- Testing requirements
- Documentation standards

## 🛠️ Common Tasks

### Task: Add New Feature
1. Read relevant sprint plan in `.agent/planning/sprints/`
2. Check architecture in `.agent/context/architecture.md`
3. Review coding standards in `.agent/context/coding-standards.md`
4. Implement following the patterns in existing code
5. Update documentation if needed

### Task: Fix Bug
1. Understand the system architecture first
2. Locate the affected component
3. Review related code and tests
4. Fix and add regression tests
5. Update changelog if significant

### Task: Refactor Code
1. Check existing refactoring docs in `.agent/services/*/refactoring/`
2. Document the refactoring plan
3. Follow the refactoring patterns
4. Update architecture docs if needed

## 📚 Key Documents by Topic

### Architecture & Design
- **Architecture Overview**: `.agent/context/architecture.md`
- **System Design**: `.agent/planning/PLAN.md`
- **Refactoring Plans**: `.agent/services/config-server/refactoring/`

### Development
- **Coding Standards**: `.agent/context/coding-standards.md`
- **Development Guide**: `docs/en/DEVELOPMENT.md` (user-facing)
- **API Documentation**: `docs/en/API.md` (user-facing)

### Project Management
- **Master Plan**: `.agent/planning/PLAN.md`
- **Sprint Plans**: `.agent/planning/sprints/sprint*.md`
- **Task Breakdowns**: `.agent/planning/tasks/`

### Service-Specific
- **Config Server**: `.agent/services/config-server/`
  - Agent guide, refactoring docs, sprint plans

## 🔍 Finding Information

### "Where is the error handling pattern?"
→ `.agent/context/coding-standards.md` (for patterns)  
→ `services/config-server/internal/errors/` (for implementation)

### "What's the current sprint objective?"
→ `.agent/planning/sprints/sprint*.md` (find the latest sprint)

### "How does the target-group relationship work?"
→ `.agent/services/config-server/refactoring/target-group-relationship.md`

### "What's the project goal?"
→ `.agent/planning/PLAN.md` or `.agent/context/project-overview.md`

### "How do I run tests?"
→ `docs/en/DEVELOPMENT.md` (user-facing documentation)

## 🚫 What NOT to Do

1. **Don't** read user documentation (`docs/`) for agent tasks
   - User docs are for humans, not comprehensive for agents
   - Agent docs in `.agent/` have more context

2. **Don't** modify sprint plans without context
   - Sprint plans are historical records
   - Only update current sprint if instructed

3. **Don't** skip reading architecture docs
   - Understanding the system prevents mistakes
   - Architecture guides implementation decisions

4. **Don't** ignore coding standards
   - Consistency is critical for maintainability
   - Follow established patterns

## 💡 Best Practices for Agents

1. **Always start with context**
   - Read `.agent/context/` before coding
   - Understand the "why" not just the "what"

2. **Follow existing patterns**
   - Look at similar existing code
   - Maintain consistency

3. **Document decisions**
   - Update `.agent/services/*/` docs for significant decisions
   - Add comments for complex logic

4. **Test thoroughly**
   - Write tests for new code
   - Run existing tests before committing

5. **Ask when uncertain**
   - Better to clarify than assume
   - Reference specific docs in questions

---

## 🔗 Quick Links

- **Agent Documentation**: [.agent/README.md](.agent/README.md)
- **User Documentation**: [docs/README.md](docs/README.md)
- **Main README**: [README.md](README.md)
- **Config Server**: [services/config-server/README.md](services/config-server/README.md)

---

**Last Updated**: 2025-12-29  
**Maintained By**: AI Coding Agents & Project Team

For human developers, start with [README.md](README.md) instead.
