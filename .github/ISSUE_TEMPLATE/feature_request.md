---
name: Feature request
about: Suggest an idea for this project
title: ""
labels: enhancement
assignees: ""
---

**What problem does this solve?**

**Proposed approach**

**Does this cross a service boundary?**

If it needs `accounts`, `admins`, or `search` to know something new
about another service's data, describe it as an event (subject +
payload) rather than a direct call -- see
[docs/architecture.md](../../docs/architecture.md) for why.
