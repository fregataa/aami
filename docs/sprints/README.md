# AAMI Sprint Planning

This directory contains sprint planning documents for the AAMI project.

## Sprint Overview

| Phase | Name | Duration | Goal | Status |
|-------|------|----------|------|--------|
| [Phase 1](./phase-1-mvp.md) | MVP | 4-6 weeks | Core functionality, 30-minute installation | 🔄 Planned |
| [Phase 2](./phase-2-enhancement.md) | Enhancement | 3-4 weeks | Differentiation features, UX improvement | ⏳ Waiting |
| [Phase 3](./phase-3-scale.md) | Scale | As needed | Large-scale environment support, ecosystem integration | ⏳ Waiting |

## MVP Completion Criteria

```bash
# MVP is complete when this works
aami init
aami nodes add --file hosts.txt --user root --key ~/.ssh/id_rsa
aami alerts apply-preset gpu-production
aami status
# → GPU monitoring working in Prometheus + Grafana
# → Xid error interpretation available
```

## Document Structure

Each sprint document includes:

- **Goals**: What should be achieved by sprint completion
- **Tasks**: Specific work items (Epic → Story → Task)
- **Acceptance Criteria**: Acceptance conditions for each feature
- **Technical Decisions**: Technical choices needed during implementation
- **Risks**: Expected issues and mitigation strategies

## Status Legend

- ⬜ Not Started
- 🔄 In Progress
- ✅ Completed
- ⏸️ On Hold
- ❌ Cancelled
