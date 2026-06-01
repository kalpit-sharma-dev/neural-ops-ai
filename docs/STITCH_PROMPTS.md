# Google Stitch Prompts — NeuralOps UI

**Purpose:** Ready-to-paste prompts to recreate the NeuralOps frontend in [Google Stitch](https://stitch.withgoogle.com/) for design exploration. Colors, fonts, and layouts match the live codebase (`frontend/src/styles/design-tokens.css`, `components/`, `pages/`).

**How to use**
1. Paste the **master design-system prompt** first.
2. Generate **one screen at a time** in a new generation, each referencing _"NeuralOps design system, dark Aurora theme"_ so styling stays consistent.
3. Get the dark theme right, then request the **light Lumen variant**.
4. Bring the export/screenshot back to the repo — it will be mapped onto `design-tokens.css`, the `ui/*` components, and existing page layouts.

**Related:** [UI_IMPROVEMENT_ROADMAP.md](./UI_IMPROVEMENT_ROADMAP.md), [FRONTEND_STACK.md](./FRONTEND_STACK.md)

---

## 1. Master design system (paste this first)

```
Design a modern enterprise AI observability platform called "NeuralOps" — a Datadog/Dynatrace-class SaaS dashboard for logs, metrics, traces, and incidents. Create a cohesive design system with a dark theme (primary) and a light theme.

BRAND & MOOD: Premium, data-dense but calm, engineering-grade. Glassmorphism accents, soft glows, subtle gradients. Think "command center for SREs."

TYPOGRAPHY:
- Display/headings: Plus Jakarta Sans (weights 600–800)
- Body/UI: Inter (400–600)
- Code, IDs, log lines: JetBrains Mono

DARK THEME "Aurora" (default):
- Backgrounds: base #09090F, surface #111118, elevated #18181F, highlight #22222C
- Text: primary #FAFAFA, secondary #A1A1AA, muted #71717A
- Accent primary cyan #22D3EE, secondary violet #A78BFA
- Accent gradient: 135deg cyan #22D3EE → violet #A78BFA → pink #F472B6
- Borders: subtle rgba(255,255,255,0.06)
- Soft radial gradient "orbs" in the page background (cyan, violet, faint pink)
- Cards have soft shadow + 1px subtle border; accent buttons glow cyan on hover

LIGHT THEME "Lumen":
- Backgrounds: base #F4F6FA, surface #FFFFFF, elevated #F8FAFC
- Text: primary #0F172A, secondary #475569, muted #94A3B8
- Accent primary teal #0891B2, secondary purple #7C3AED
- Clean paper feel, soft shadows, no heavy glows

SEMANTIC / SEVERITY COLORS (dark):
- Critical #FB7185, P1 #FB923C, P2 #FBBF24, P3 #38BDF8, P4 gray #A1A1AA
- Healthy #34D399, Degraded #FBBF24, Down #FB7185
- Each used as a small pill/badge with a 14%-opacity tinted background

SHAPE & SPACING:
- Corner radii: 8px (controls), 12px (cards), 16px (panels)
- 4/8/12/16/24/32px spacing scale
- Chart series palette: #22D3EE, #A78BFA, #FBBF24, #FB7185

CORE COMPONENTS to define: top bar, left sidebar nav, KPI metric card (with label, big value, trend %, mini sparkline), data card/panel, severity badge/pill, status dot (colored, with pulse for active), table rows, search input with "⌘K" shortcut hint, empty state (icon + title + description + button), buttons (primary gradient, secondary outline, ghost), toast notifications.
```

---

## 2. App shell

```
Design the main application shell for NeuralOps using the established design system.

LAYOUT:
- Left sidebar (252px, collapsible to 68px): brand mark "N" + "NeuralOps" wordmark at top. Nav grouped into collapsible sections with small uppercase labels: OBSERVE (Command Center, Log Explorer, Trace Explorer, Service Flow, Metrics, Dashboards, Service Map), RESPOND (Incidents, Alerts, SLOs), PLATFORM (Infrastructure, Kubernetes, Databases, RUM, Cloud, Security), AUTOMATE (AI Assistant, Workflows, Notebooks), ADMIN (Marketplace, Integrations, Settings). Each item: icon + label. Active item has cyan tint + left accent. Items show a star "pin" on hover; pinned favorites appear in a "Favorites" group at top.
- Top bar (60px): hamburger (mobile), brand, center cluster with Environment dropdown (ALL/PROD/STAGING/DEV) + Time range dropdown (1h/6h/24h/7d) + a read-only search box "Search… (⌘K)", right side: theme toggle (sun/moon/monitor), bell with unread badge, live status dot, user avatar + name + tenant, logout.
- Main content area fills the rest, scrollable.
- Include a thin demo banner above the top bar.
Show the dark Aurora theme.
```

---

## 3. Command Center (dashboard)

```
Design the "Command Center" dashboard screen for NeuralOps (dark Aurora theme), inside the app shell.

CONTENT TOP-TO-BOTTOM:
- Page header "Command Center" + subtitle "Real-time observability overview". Right side: "Updated 12s ago" text, a Refresh button, and a red "War room" button (only when a P1 is open).
- Incident strip (only if critical): P1/P2 count badges + a horizontal list of latest incident titles as links.
- Row of 4 clickable KPI cards: Active Incidents (with P1–P4 mini badges), Error Rate 1h (with trend % and sparkline), Avg Latency p99 (sparkline), MTTR Today. Use the metric-card component.
- Two-column row: "Incident Timeline (24h)" card with colored horizontal bars by severity; "Top Failing Services" card listing services with a status dot, a thin error bar, and error count.
- Full-width "Error Rate by Service" stacked area chart (cyan/violet/amber/pink series) with dark grid.
- Three-column row: "Active Anomalies" (service + message + warning badge), "Recent Deployments", "AI Insights" (bulleted insights + timestamp).
Data-dense but breathable. Cards use subtle borders and soft shadows; faint gradient orbs in the background.
```

---

## 4. Log Explorer (core screen)

```
Design the "Log Explorer" screen for NeuralOps (dark Aurora theme), inside the app shell. This is a 3-column log analysis workspace.

LEFT (300px) — Filters sidebar: collapsible facets for Services (with counts + health dot), Severity (FATAL/ERROR/WARN/INFO with counts), Environment, Host, Pod, and toggles "Has stack trace", "Has AI explanation", "My services only".

CENTER — Log stream:
- Search bar with placeholder "Search logs… or try 'payment failures after 10pm' (AI)" and a "⌘K" hint.
- Mode toggle buttons: Text / Regex / AI/NLP. A "Live tail" checkbox. A row of saved-search chips with delete buttons + a "Save" control.
- A small bar histogram of log volume over time (clickable to narrow the range).
- A virtualized log table: each row has a severity letter badge, relative time, clickable service name, the log message (with highlighted matches), a short trace ID link, and a sparkle icon when an AI explanation exists. Color the left edge of each row by severity.
- Footer: "12,438 results · 23ms · live" plus Export JSON / Export CSV / Share buttons.

RIGHT (400px, slide-in) — Log detail panel: full message, structured fields, trace context, and an "AI explanation" section in plain English.
Show a realistic banking/payments dataset (services like payment-api, upi-service, auth-service, ledger-service).
```

---

## 5. Incident Detail (focus mode)

```
Design the "Incident Detail" focus screen for NeuralOps (dark Aurora theme) — full-page, no sidebar.

- Header: severity badge (P1 orange), incident title, status, started-time, affected services as chips.
- Tabs: Overview, Timeline, Logs, Traces, Evidence.
- Overview: summary card + AI root-cause analysis card with confidence.
- Timeline: vertical event list with per-event-type icons and relative + absolute timestamps.
- A sticky action bar pinned to the bottom: Assignee input, resolution Notes textarea, and buttons Acknowledge / Assign / Resolve.
Emphasize fast on-call response. Use severity color accents and glassmorphism on the sticky bar.
```

---

## 6. Trace Explorer & Trace Detail

```
Design the "Trace Explorer" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Search/filter bar: service dropdown, operation, min duration, status (OK/ERROR), time range.
- A latency heatmap (service rows × time columns, color from cyan = fast to pink = slow).
- A results table of traces: trace ID (mono), root service, operation, duration, span count, status badge, start time. Rows clickable.
- "Compare traces" affordance: select two traces to open a side-by-side compare view.

Then a "Trace Detail" view:
- Header: trace ID, total duration, span count, service, status.
- A waterfall of spans: indented rows with service name, operation, a horizontal duration bar colored by service, expand/collapse carets, a "copy span ID" button, and a "View logs" link (deep-links to logs filtered by traceId).
- A toggle to a flame graph with zoom in/out/reset controls.
```

---

## 7. Service Map / Topology

```
Design the "Service Map" topology screen for NeuralOps (dark Aurora theme), inside the app shell.

- Full-canvas interactive node graph: service nodes as rounded cards showing name, health status dot, error rate %, and throughput. Edges are directional with thickness by call volume and color by error rate.
- Zoom and pan controls bottom-right.
- A floating legend (bottom-left, glass panel): health colors (healthy green, degraded amber, down pink) and dependency types.
- Top filter bar: namespace, health, dependency-type filters.
- Clicking a node opens a right-side detail panel: service overview, error rate sparkline, top dependencies, and links to its entity page / traces / logs.
```

---

## 8. Metrics Explorer

```
Design the "Metrics Explorer" screen for NeuralOps (dark Aurora theme), inside the app shell.

- A PromQL query editor: monospace textarea with autocomplete suggestions dropdown, a query history list, inline error underline, and a "Run (Ctrl+Enter)" hint.
- Chart-type toggle: Line / Bar / Stat / Table.
- A large result chart area using the cyan/violet/amber/pink series palette on a dark grid, with a legend and hover tooltip.
- A metric catalog sidebar: searchable list of metric names with units and labels.
```

---

## 9. Alerts

```
Design the "Alerts" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Tabs: Active, Rules, Channels, Silences, Escalation, History.
- Active tab: alerts grouped by service (collapsible group headers with counts). Each alert row: severity badge, title, service, fired-time, and a checkbox for bulk-acknowledge; a bulk "Acknowledge selected" action bar appears when rows are selected.
- Rules tab: a form using labeled inputs (Rule name, Service pattern, Source, Severity dropdown, Enabled toggle) with inline validation, plus a list of existing rules with edit/delete.
- Silences tab: create-silence form (Service pattern, Reason, Duration) with a live "this would affect N alerts" preview count.
Use consistent labeled form fields with hints and error text.
```

---

## 10. AI Assistant (chat)

```
Design the "AI Assistant" chat screen for NeuralOps (dark Aurora theme), inside the app shell.

- Centered conversation column. Assistant messages rendered as rich markdown; each answer can include a "Sources" section with citation chips that link to log lines, traces, and incidents.
- Page-aware suggested prompt chips at the top (e.g. "Summarize this incident", "Which service has worst latency?", "Show failed UPI transactions today").
- A streaming indicator with status text ("Searching logs…", "Analyzing patterns…", "Generating response…").
- Bottom input bar with a send button and a stop/square button while streaming.
A subtle Cpu/AI glyph as the assistant avatar. Cyan→violet gradient accents.
```

---

## 11. RUM, Kubernetes, Infrastructure, Databases

```
Design four platform monitoring screens for NeuralOps (dark Aurora theme), inside the app shell, consistent with the design system:

A) RUM: Core Web Vitals cards (LCP, FID, CLS) with distribution charts, plus a sessions table (user, page, device, country, duration, errors) where each row links to a session replay player (play/pause, timeline scrubber, event list).

B) Kubernetes: a sortable workload table (deployments/pods) with namespace, ready/desired, restarts, CPU, memory, status badge.

C) Infrastructure: a host inventory list; clicking a host opens a right-side drawer with CPU / memory / disk sparklines and big current-value stats.

D) Databases: a slow-query table sortable by p95 latency, with query text (mono, truncated), call count, and a link to related traces.
```

---

## 12. Light theme variant

```
Now produce the light "Lumen" variant of the [SCREEN NAME] using the light palette: base #F4F6FA, surface #FFFFFF, elevated #F8FAFC, text primary #0F172A / secondary #475569 / muted #94A3B8, accent teal #0891B2, secondary purple #7C3AED. Clean paper feel, soft shadows, no heavy glows. Keep the same layout and components as the dark version.
```

---

## Integration notes (for bringing designs back into the repo)

| Stitch concept | Maps to in repo |
|----------------|-----------------|
| Color tokens | `frontend/src/styles/design-tokens.css` (`[data-theme='dark'|'light']`) |
| Buttons / badges / cards | `frontend/src/components/ui/*` |
| Sidebar / top bar | `frontend/src/components/layout/*`, `lib/navConfig.ts` |
| Charts | Recharts + `frontend/src/lib/chartTheme.ts`, `chartColors.ts` |
| Page screens | `frontend/src/pages/*` |
| Empty states | `components/ui/DomainEmptyState.tsx` |

Keep new visuals expressed as **design tokens + existing components** rather than one-off styles, so both Aurora and Lumen themes stay in sync.
