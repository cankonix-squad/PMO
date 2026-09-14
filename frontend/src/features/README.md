# Feature boundaries

Project feature implementation lives in `projects/`:

- `list/ProjectsPage.tsx`: project list and selected-project layout.
- `list/components`: project form, selected-project panel, and list UI.
- `list/hooks`: list queries, selection state, and project mutations.
- `detail/ProjectDetailPage.tsx`: detail-page layout and panel composition.
- `detail/components`: task, milestone, issue, risk, budget, contract, document,
  corrective-action, baseline, and snapshot panels/forms.
- `detail/hooks`: page state/form coordination, query projections, and separate
  mutation hooks per resource.
- `list/schemas.ts`, `detail/schemas.ts`: validation schemas, form types/constants.
- `list/utils.ts`, `detail/utils.ts`: form readers, payload conversion, formatting.
- `components/`: project-only Gantt, inspection, health, control, periodic reports.

Next.js route files are thin entry points. Shared application UI stays under
`src/components`. The existing `src/services` API layer and `src/types` contracts
remain shared so other pages can consume project data without importing feature
implementation details. Preserve query keys, mutation payloads, and invalidation
behavior when editing hooks.

Run `npm run type-check` and `npm run lint`. Run `npm run build` in an isolated
checkout or after stopping the dev server. Browser regression checks use mocked
API data and are documented in `tests/README.md`.
