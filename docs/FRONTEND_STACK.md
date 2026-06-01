# Frontend Stack (Phase 9 — completed)

| Component | Implementation |
|-----------|----------------|
| **Tailwind CSS v4** | `@tailwindcss/vite` + `@import "tailwindcss"` in `global.css` (alongside design tokens) |
| **TanStack Router** | Route tree in `src/router.tsx`; `RouterProvider` in `App.tsx` |
| **Data fetching** | TanStack Query |
| **UI primitives** | Radix + custom `components/ui/*` |

Legacy `react-router-dom` has been removed from the runtime path. Custom CSS tokens remain the primary styling layer; Tailwind utilities are available for incremental adoption.
