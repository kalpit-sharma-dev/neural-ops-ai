# Google Stitch Prompts — NeuralOps UI

**Purpose:** Ready-to-paste prompts to recreate the NeuralOps frontend in [Google Stitch](https://stitch.withgoogle.com/) for design exploration. Colors, fonts, and layouts match the live codebase (`frontend/src/styles/design-tokens.css`, `components/`, `pages/`).

**How to use**
1. Paste the **master design-system prompt** first.
2. Generate **one screen at a time** in a new generation, each referencing _"NeuralOps design system, dark Aurora theme"_ so styling stays consistent.
3. Get the dark theme right, then request the **light Lumen variant**.
4. Bring the export/screenshot back to the repo — it will be mapped onto `design-tokens.css`, the `ui/*` components, and existing page layouts.

**Related:** [STITCH_IMPLEMENTATION.md](./STITCH_IMPLEMENTATION.md) (phase-wise code checklist), [UI_IMPROVEMENT_ROADMAP.md](./UI_IMPROVEMENT_ROADMAP.md), [FRONTEND_STACK.md](./FRONTEND_STACK.md), [stitch-screens.json](./stitch-screens.json) (screen IDs for your project)

**Your project ID:** `37487902246048120` — title: *Prompt-Based Design Generator*

---

## Screen inventory (your Stitch project)

### Already designed (desktop) — use as visual reference for new screens

| Screen | node-id | Notes |
|--------|---------|--------|
| Command Center | `9219f212f70f4257a2d51df33852d7c1` | Full shell, War room, KPIs, charts |
| Log Explorer | `0cd0afabbace4bfebac53ffee97b147f` | 3-pane, AI root cause, saved views |
| Trace Explorer | `c7cf6a4896f440d1aca3628b88348c28` | **Use desktop** — not mobile frame |
| Incident Detail | `6f46e7ed41e14e6da842d05619cc6e4a` | Focus mode, AI card, runbook CTA |
| Alerts | `6cc6a9e5a64b41c681d650117ef1c022` | Grouped by service, bulk actions |
| Metrics Explorer | `1637ab0178844d77a570950f338289e1` | PromQL + catalog sidebar |
| Service Map | `030a5837ecb84808bac323d7ec92029a` | Topology + right drawer |
| AI Assistant | `4cb39d9f61af4865b830ef7e326105dd` | Citations, prompt chips |
| Infrastructure & K8s | `ecc529cb841540489bc2d8dbebf640c5` | Workloads + host drawer |

Preview URL pattern: `https://stitch.withgoogle.com/preview/37487902246048120?node-id={node-id}`

### Phase 1–2 complete (§13–§26) — canonical desktop frames

| Screen | node-id |
|--------|---------|
| Login | `a864750f547e413bb5b63d4e3a52a181` |
| Command Center | `9219f212f70f4257a2d51df33852d7c1` |
| Log Explorer | `0cd0afabbace4bfebac53ffee97b147f` |
| Trace Explorer | `c7cf6a4896f440d1aca3628b88348c28` |
| Service Flow | `7c7533d543bf411d8a1410c099990804` |
| Incidents List | `c2d41efbe7bf4e0597afb9e7a0bbd1dd` |
| SLO Management | `df42bd8ec62540f4ab653562486b0054` |
| Dashboards List / Editor | `fffd2350…` / `c770f81c…` |
| Workflows Editor | `d980e4d6b6ba4263b68aa8d3645f819e` |
| Notebooks | `654c87d65f394766a4188fc74156f8d8` |
| Transaction Journey | `da8a743169614b0788ee1351fb563b85` |
| Anomaly Detection Dashboard | `4c4064947cf54909be39bf166c24d446` |
| Settings: API Keys / SSO | `48c7987f…` / `c1848c5c…` |
| RUM | `d201f47f19d94ada8406c28bd972153b` |
| Synthetic Monitoring | `3c6d7ae4ea3b4f07aa5f971befdcfca4` |
| Security Monitoring | `1072b670f8004514b454252c3773831e` |
| Lumen (Command Center, Logs, Login) | `c061bc55…` / `7d3f5202…` / `e3de9f36…` |

Full list: [stitch-screens.json](./stitch-screens.json) (84 screens as of last API sync).

**Implement in code:** follow [STITCH_IMPLEMENTATION.md](./STITCH_IMPLEMENTATION.md) (Phases 0–8).

### Build next in Stitch — Phase 3 (§27–§39) — complete

| § | Screen | Maps to route |
|---|--------|----------------|
| **27** | Databases | `/databases` |
| **28** | Middleware | `/middleware` |
| **29** | Cloud monitoring | `/cloud` |
| **30** | Integrations | `/integrations` |
| **31** | Marketplace | `/marketplace` |
| **32** | Settings hub | `/settings` |
| **33** | Settings — Users, Audit, Usage | `/settings/users`, `/settings/audit`, `/settings/usage` |
| **34** | Trace detail | `/traces/$traceId` |
| **35** | Trace compare | `/traces/compare` |
| **36** | Entity page | `/entities/$type/$id` |
| **37** | Log settings + Trace settings | `/logs/settings`, `/traces/settings` |
| **38** | Settings — Policies & On-call | `/settings/policies`, `/settings/oncall` |
| **39** | Lumen extended batch | key remaining pages |

**Consistency rule for every new prompt:** Add this line at the end:

> Match the exact sidebar, env tabs (PROD/STAGING/DEV), teal accent, card style, and typography from my existing **NeuralOps Command Center** screen in this project. Desktop 2560×2048 or 3040×2048.

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

## Phase 2 — New screens (paste after master prompt)

### 13. Login & auth (P1)

```
Design a full-screen Login page for NeuralOps (dark Aurora theme) — NO app sidebar.

- Centered glass card on gradient orb background (cyan/violet subtle glow).
- NeuralOps wordmark + "AI Enterprise Control" subtitle.
- Email field, "Continue" primary button, divider "or".
- "Sign in with SSO" secondary button (Google/OIDC style).
- Small footer: "Demo: demo@neuralops.ai" hint for local dev.
- Top-right: theme toggle (sun/moon).
- Optional second frame: SSO redirect / "Signing you in…" loading state with spinner.
Premium, trustworthy, not playful. Match banking/SRE audience.
Match the exact teal accent and typography from my existing NeuralOps Command Center screen in this project. Desktop 1440×900.
```

---

### 14. Incidents list (P1)

```
Design the "Incidents" list screen for NeuralOps (dark Aurora theme), inside the full app shell (same sidebar as Command Center: Observe / Respond / Infra / Intelligence sections, user footer).

- Page title "Incidents" + subtitle "Active and recent operational incidents".
- Filter row: search ("Search incidents… /"), severity multi-select (P1–P4), status (Open/Investigating/Resolved), service dropdown, time range.
- Table: Severity pill | Title (link) | Status | Affected services (chips) | Owner | Started | Duration.
- Row hover highlight; P1 rows have subtle red left border.
- Empty state: "No incidents — all systems normal" with illustration.
- Top-right: "+ Create incident" ghost button (optional).
Show 8–10 realistic rows (payment-api latency, UPI timeout, auth-service deployment).
Match existing Command Center shell exactly. Desktop 2560×2048.
```

---

### 15. Service Flow (P1)

```
Design the "Service Flow" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Header: "Service Flow" + subtitle "Request path across microservices".
- Horizontal DAG / left-to-right flow diagram: nodes = services (payment-api → auth-service → ledger-service → stripe-gateway), edges show call count, error rate %, p95 latency on hover.
- Nodes are clickable cards with health dot (green/amber/red).
- Top filters: entry service, time range, min error rate.
- Bottom mini-panel: selected edge details (source → target, throughput, errors).
- "Compare with baseline" toggle (24h ago).
Alternative: sankey-style flow for high traffic paths.
Match Command Center sidebar and env tabs. Desktop 3040×2048.
```

---

### 16. SLOs (P2)

```
Design the "SLOs" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Tabs: Overview | Objectives | Error budgets | History.
- Overview: grid of SLO cards — each shows service name, objective (99.9% availability), current % (big number), error budget remaining (progress bar green/amber/red), burn rate sparkline.
- "+ Create SLO" primary button opens a slide-over form: name, service, SLI type (availability/latency), target %, window (7d/30d), alert thresholds.
- Table view toggle for power users.
- One SLO in "breach" state (red) for payment-api.
Match existing NeuralOps design system. Desktop 2560×2048.
```

---

### 17. Custom Dashboards (P2)

```
Design the "Dashboards" screen for NeuralOps (dark Aurora theme), inside the app shell.

A) List view: searchable grid of dashboard cards (name, owner, shared badge, last edited, thumbnail preview). "+ New dashboard" CTA.

B) Editor view (second artboard): 12-column grid with drag handles on tiles. Tile types: time series chart, stat single number, table, logs panel, markdown text. Top bar: dashboard name editable, variables dropdown (service, env), Save, Share, TV mode.
Show 4–6 tiles populated with payment/upi metrics.
Match Command Center visual style. Desktop 3040×2048.
```

---

### 18. Workflows editor (P2)

```
Design the "Workflows" editor for NeuralOps (dark Aurora theme), inside the app shell.

- Left: list of workflows (name, trigger type, enabled toggle, last run status).
- Center: visual step pipeline — cards connected by arrows: Trigger (Alert fired) → Condition (severity = P1) → Action (Create incident) → Action (Notify Slack) → Action (Run runbook).
- Each step card: icon, title, edit menu; drag reorder handles.
- Right inspector: configure selected step (webhook URL, channel, message template).
- Top: workflow name, Enable switch, "Test run" button.
- Bottom log: last 3 executions with success/fail badges.
Not a raw JSON editor — visual first. Desktop 2560×2048.
```

---

### 19. Notebooks (P2)

```
Design the "Notebooks" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Left sidebar: notebook list + "New notebook".
- Main: stacked cells like Jupyter — each cell has type pill (Markdown | Log query | PromQL | Python), run button, output area below.
- Show one markdown cell (title), one PromQL cell with chart output inline, one log query cell with table output.
- Top bar: notebook name, Run all, Schedule, Share.
- Cell add menu: + Markdown / + Query / + Chart.
Desktop 2560×2048. Match NeuralOps tokens.
```

---

### 20. Transaction Journey — UPI / banking (P3)

```
Design the "Transaction Journey" screen for NeuralOps (dark Aurora theme), inside the app shell. Banking/UPI focus.

- Search by txn ID, UPI VPA, or amount range.
- Swimlane timeline: User → Mobile app → API Gateway → payment-api → NPCI switch → Beneficiary bank — each step shows status (success/fail), latency, and link to trace/logs.
- Right panel: txn metadata (amount INR, VPA, MCC, device, geo).
- Highlight one failed UPI txn with red step "NPCI timeout".
Desktop 2560×2048.
```

---

### 21. Anomaly Detection (P3)

```
Design the "Anomaly Detection" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Feed of detected anomalies: timestamp, service, metric, deviation score, severity, "Investigate" button.
- Top chart: anomaly score over time with shaded anomaly regions.
- Filters: service, metric family, confidence threshold.
- Side panel on click: AI explanation, related incidents, link to metrics/logs.
Desktop 2560×2048.
```

---

### 22. Settings — API keys & SSO (P3)

```
Design "Settings" for NeuralOps (dark Aurora theme), inside the app shell with Settings highlighted in Admin section.

Use left sub-nav: Profile | Users | API Keys | SSO | Audit | Usage.

Focus artboard on API Keys:
- Table: name, prefix (sk_live_••••), scopes, created, last used, revoke.
- "+ Create API key" opens modal with name, scopes checkboxes, expiry — show secret once with copy button.

Second artboard: SSO — OIDC provider URL, client ID (masked), enabled toggle, "Test connection" button.
Form fields with labels, hints, inline errors. Desktop 2560×2048.
```

---

### 23. RUM — Session replay (P4)

```
Design "RUM" for NeuralOps (dark Aurora theme), inside the app shell.

- Vitals row: LCP, INP/FID, CLS with distribution mini-charts.
- Sessions table: session ID, user, page, device, country flag, duration, errors, replay icon.
- Session replay view (full width): video-style player, timeline scrubber, event list (click, navigation, error), console errors tab.
Desktop 2560×2048. Match existing NeuralOps screens.
```

---

### 24. Synthetic monitoring (P4)

```
Design "Synthetic Monitoring" for NeuralOps (dark Aurora theme), inside the app shell.

- Monitor list: name, URL/endpoint, locations, frequency, last status (up/down), p95 latency.
- Geo map or horizontal bar chart: latency by location (Mumbai, Singapore, Virginia).
- Detail: run history table, waterfall of last check, screenshot thumbnail.
Desktop 2560×2048.
```

---

### 25. Security (P4)

```
Design "Security" for NeuralOps (dark Aurora theme), inside the app shell.

- Summary KPIs: attacks blocked 24h, top threat type, top targeted service.
- Table: time, attack type (SQLi, XSS, brute force), source IP, service, action (blocked/alerted).
- Click row → detail drawer with payload snippet (mono), geo, linked traces, rule that matched.
Desktop 2560×2048.
```

---

### 26. Lumen light theme — batch (P4)

Run **after** dark screens are approved. One prompt per screen:

```
Recreate my existing NeuralOps [SCREEN NAME] artboard in light "Lumen" theme only.
Palette: base #F4F6FA, surface #FFFFFF, text #0F172A, accent #0891B2, purple #7C3AED.
Same layout and components as the dark version — no layout changes.
Screens to do first: Command Center, then Log Explorer, then Login.
```

---

## Phase 3 — remaining app screens (§27–§39)

Use the **consistency rule** at the end of every prompt below. Generate **one screen per Stitch generation** unless the prompt explicitly asks for a second artboard.

---

### 27. Databases (P3)

```
Design the "Databases" screen for NeuralOps (dark Aurora theme), inside the app shell (PLATFORM section: Databases highlighted).

- Page title "Databases" + subtitle "Query performance and connection pools".
- Top filter row: engine type (PostgreSQL, MySQL, Redis), environment, search by instance name.
- Grid of database instance cards (3–4 columns): instance name, engine badge (Postgres/MySQL), health dot, KPIs — QPS, slow queries count, active connections, pool utilization % mini bar.
- One card selected/highlighted with cyan border.
- Bottom panel (full width): "Slow queries" sortable table — Query (mono, truncated), Calls, Avg ms, P95 ms, Last seen, link "View trace".
- Show 6–8 realistic payment/upi queries (SELECT … FROM transactions, UPDATE ledger…).
- Empty state on panel if no instance selected: "Select a database to inspect slow queries".
Desktop 2560×2048. Match Command Center shell exactly.
```

---

### 28. Middleware — Kafka lag (P3)

```
Design the "Middleware" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Page title "Middleware" + subtitle "Kafka consumer lag and queue depth".
- Summary KPI row: total lag (messages), topics in alert, oldest consumer group age.
- Main table: Topic | Consumer group | Partition | Lag | Lag trend (sparkline) | Status pill (OK / Warning / Critical).
- Highlight 2 rows in amber/red (payment-events lag 12.4k, upi-notifications lag 890).
- Optional second card: "RabbitMQ queues" placeholder with muted "Coming soon" or show 3 queue rows for parity.
- Top-right: refresh + export CSV ghost buttons.
Desktop 2560×2048.
```

---

### 29. Cloud monitoring (P3)

```
Design the "Cloud monitoring" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Page title "Cloud monitoring" + subtitle "AWS, Azure, and GCP observability dashboards".
- Provider tabs: AWS | AZURE | GCP (AWS active).
- For active provider: 2–3 dashboard cards (e.g. "EC2 — ap-south-1", "RDS — payment-db", "Lambda — webhook-handler") each with region label and a row of metric chips (CPUUtilization, NetworkIn, ErrorRate) — one chip selected/cyan.
- Below selected metric: time series chart (24h) with area fill, legend, hover tooltip.
- Right sidebar mini-panel: linked resources count, last sync time, "Open in provider console" link.
Show realistic AWS metrics for a fintech stack. Desktop 2560×2048.
```

---

### 30. Integrations (P3)

```
Design the "Integrations" screen for NeuralOps (dark Aurora theme), inside the app shell (ADMIN → Integrations highlighted).

- Page title "Integrations" + subtitle "Connect incident and observability tools".
- Grid of integration tiles (2–3 columns): Jira, Slack, PagerDuty, ServiceNow, AWS, Azure, GCP — each with logo placeholder, description, Connected/Available badge, primary "Connect" or "Configure" button.
- Jira tile shows "Connected" green badge; Slack "Available".
- Modal overlay (centered) for "Connect Slack": OAuth client ID, client secret (masked), webhook URL optional, Cancel + Connect buttons, inline validation hint on empty required field.
- Footer note: "OAuth redirects use your tenant SSO domain".
Desktop 2560×2048.
```

---

### 31. Marketplace (P3)

```
Design the "Marketplace" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Page title "Marketplace" + subtitle "Extensions and integration catalog".
- Top-right link: "Manage connections →".
- Card grid: extension tiles — Grafana panels (installed), Jira incident sync (available), PagerDuty on-call (installed), plus 2–3 more (OpenTelemetry collector, ServiceNow CMDB).
- Each card: icon, name, type pill (visualization | integration), status badge (installed / available), short description, CTA "Install" or "Configure".
- Filter chips: All | Installed | Available | Integrations.
Desktop 2560×2048.
```

---

### 32. Settings hub (P3)

```
Design the "Settings" overview screen for NeuralOps (dark Aurora theme), inside the app shell (ADMIN → Settings highlighted).

- Page title "Settings" + subtitle "Platform configuration and administration".
- Card grid (3 columns, no left sub-nav on this page — hub only): Users & Teams, API Keys, Audit Log, Usage, Log Settings, Integrations, SSO / IdP, Tenant policies, On-call schedules, Design System, System Health.
- Each card: title, muted one-line description, "Open →" link style CTA.
- Do NOT duplicate the API Keys or SSO detail layouts — this is navigation hub only.
Desktop 2560×2048.
```

---

### 33. Settings — Users, Audit, Usage (P3)

Three artboards in one generation (or one at a time):

**A) Users** (`/settings/users`):
- Inside app shell with Settings context; breadcrumb Settings → Users.
- Table: Email | Role (Admin/Editor/Viewer badges) | Status (Active/Disabled) | Last login.
- Top-right: "+ Invite user" primary button.

**B) Audit log** (`/settings/audit`):
- Table: Timestamp | Actor | Action | Resource | IP (mono).
- Filters: date range, action type, actor search.
- Show 10 realistic audit rows (api_key.created, sso.updated, user.invited).

**C) Usage** (`/settings/usage`):
- KPI cards: Logs ingested (GB/24h), Traces (spans/24h), Metrics cardinality, AI tokens consumed.
- Stacked bar or area chart: ingest by signal type over 7 days.
- Table: top services by ingest volume.

Desktop 2560×2048 each artboard. Match Settings: API Keys visual style.
```

---

### 34. Trace detail (P3)

```
Design the "Trace detail" screen for NeuralOps (dark Aurora theme), inside the app shell (Trace Explorer context).

- Header: trace ID (mono, truncated), status badge (OK / ERROR), service name, span count, total duration ms.
- Actions: Sampling link, Compare, ← Back to explorer.
- View toggle: Waterfall | Flame graph (Waterfall active).
- Main area: horizontal waterfall / Gantt — spans as bars colored by service (payment-api cyan, auth-service violet, db gray), nested rows, duration axis in ms, error span in red.
- Right drawer or bottom panel: selected span details — operation name, tags key-value, logs link, duration breakdown.
- Below waterfall: "Continuous profiling" card with CPU flame mini-chart placeholder.
Use realistic 12–15 spans for a failed UPI checkout. Desktop 3040×2048.
```

---

### 35. Trace compare (P3)

```
Design the "Compare traces" screen for NeuralOps (dark Aurora theme), inside the app shell.

- Page title "Compare traces" + subtitle "Side-by-side waterfall comparison".
- Picker card: Trace A input | Trace B input | Compare primary button; below — "Recent traces" chips (service · duration) to fill pickers.
- Two equal columns below: Trace A (green header) and Trace B (amber header), each with mini metadata row (status, service, duration, span count).
- Stacked waterfalls in each column, aligned time axis, highlight duration delta on matching spans.
- Bottom summary bar: "Trace B is 340ms slower — bottleneck: payment-api → ledger-service".
Desktop 2560×2048.
```

---

### 36. Entity page — service (P3)

```
Design the "Entity" detail screen for NeuralOps (dark Aurora theme), inside the app shell.

Example entity: service "payment-api", health dot green.
- Header: entity type pill (SERVICE), display name, health badge, breadcrumb from Service Map.
- Tab bar: Overview | Metrics | Traces | Logs | Incidents.
- Overview tab: 3 KPI cards (Error rate %, Throughput rpm, P99 latency ms), "Quick links" row (Traces, Logs, Metrics, Service map, Incidents).
- Metrics tab (show as second artboard OR tab content): P99 latency line chart 24h.
- Traces tab preview: compact table last 5 traces with duration and status.
Use payment-api / UPI context. Desktop 2560×2048.
```

---

### 37. Log settings + Trace settings (P3)

Two artboards:

**A) Log settings** (`/logs/settings`):
- Breadcrumb Settings → Log settings.
- Two-column layout: "Log metric rules" (name, regex pattern, add form) + list of rules; "Parsing rules" (name, grok/regex, target field) + list.
- Link back to Settings hub.

**B) Trace sampling & retention** (`/traces/settings`):
- Form card: Retention (days), Head sample rate (0–1), Tail sample rate for errors (0–1), Save primary button.
- Muted help text explaining head vs tail sampling and ClickHouse TTL.

Desktop 2560×2048 each.
```

---

### 38. Settings — Policies & On-call (P3)

Two artboards:

**A) Tenant policies** (`/settings/policies`):
- Cards: Log retention (days slider), Max ingest GB/day, PII redaction toggle, allowed regions multi-select.
- Save bar sticky at bottom.

**B) On-call schedules** (`/settings/oncall`):
- Current rotation card: team name, primary/secondary, timezone, next handoff countdown.
- Calendar week view with colored shifts; "+ Add schedule" button.
- Escalation policy snippet: P1 → PagerDuty → Slack #war-room.

Desktop 2560×2048 each.
```

---

### 39. Lumen light theme — extended batch (P3)

Run **after** Phase 3 dark screens are approved. One prompt per screen:

```
Recreate my existing NeuralOps [SCREEN NAME] artboard in light "Lumen" theme only.
Palette: base #F4F6FA, surface #FFFFFF, text #0F172A, accent #0891B2, purple #7C3AED.
Same layout and components as the dark version — no layout changes.
Priority order: Incidents List → SLO Management → Dashboards List → Trace detail → Settings hub.
```

Optional second wave: Transaction Journey, RUM, Security Monitoring, Integrations.

---

## Stitch tips (faster, consistent output)

1. **Reference an existing screen** in Stitch: "Extend the same project; match Command Center sidebar."
2. **One screen per generation** — avoid 5 screens in one prompt.
3. **Desktop only** unless you explicitly need mobile (390×884).
4. **Name layers clearly** in Stitch (Sidebar, KPI Row, Chart) — helps when exporting to code.
5. After each screen, copy the preview link (`node-id=…`) into [stitch-screens.json](./stitch-screens.json) or a spreadsheet.
6. **Do not redesign** core Phase 1–2 screens unless doing a deliberate Lumen pass (§26, §39).
7. After Phase 3 generations, update [stitch-screens.json](./stitch-screens.json) with new `node-id` values.

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
