# Google Stitch → Cursor: End-to-End Integration Prompt

Paste this entire prompt into Cursor (Agent mode recommended). Customize the sections marked with `[ ]`.

---

## MASTER PROMPT

```
You are a senior frontend engineer helping me build a production-ready app.
I have designed all my screens in Google Stitch and connected the Stitch MCP
to this Cursor workspace. Your job is to:

1. Fetch every screen from Stitch
2. Analyse the design system (colors, typography, spacing, components)
3. Scaffold the full project structure
4. Implement every screen as a [React / Vue / HTML] component/page
5. Wire up routing and shared state
6. Flag any screens that need modification and apply my instructions

---

### STEP 1 — Fetch all Stitch screens

Use the Stitch MCP tools to:
- List every frame / screen in my Stitch project: [YOUR PROJECT NAME / ID]
- Export each screen as an image or structured JSON (whichever the MCP supports)
- Print a numbered inventory:  ID | Screen Name | Route suggestion

Do not write any code yet. Just output the inventory table.

---

### STEP 2 — Extract the design system

From the fetched screens, identify and output:

**Colors**
- Primary, secondary, accent, background, surface, error, success
- Map each to a CSS variable name: --color-primary, etc.

**Typography**
- Font families used
- Scale: display / heading / body / caption sizes & weights
- Line-height and letter-spacing patterns

**Spacing & Layout**
- Base unit (4px / 8px grid?)
- Common padding / margin values
- Breakpoints if responsive layouts are visible

**Component Inventory**
- List every reusable UI element seen across screens
  (buttons, inputs, cards, nav bars, modals, chips, etc.)
- Note variants (primary button, ghost button, etc.)

Output this as a structured design-tokens file I can save as
`src/styles/tokens.css` (or `tokens.js` if JS-in-CSS).

---

### STEP 3 — Propose project structure

Based on the screen inventory, suggest a folder structure:

src/
  components/       # shared/atomic components
  pages/            # one file per screen
  layouts/          # shell layouts (with nav, sidebar, etc.)
  styles/           # tokens, globals, reset
  routes/           # router config
  store/            # global state (if needed)
  assets/           # icons, images exported from Stitch

Also suggest:
- Routing library: [React Router / Vue Router / Next.js App Router / other]
- State management: [Zustand / Redux / Pinia / Context / none]
- CSS approach: [Tailwind / CSS Modules / styled-components / plain CSS]

Wait for my approval before scaffolding.

---

### STEP 4 — Scaffold and implement

After I approve the structure:

a) Create the folder structure and install dependencies.

b) Generate `tokens.css` from Step 2.

c) For EACH screen in the inventory:
   - Create the component/page file
   - Implement the UI pixel-close to the Stitch design
   - Use real, semantic HTML (no div-soup)
   - Wire up navigation links to other screens
   - Add TODO comments where real data / API calls are needed
   - Do NOT use placeholder lorem ipsum text — use realistic copy
     matching the app's purpose: [DESCRIBE YOUR APP IN 1 SENTENCE]

d) Create a router/index file that maps every route to its page.

e) Create a root App component that wraps everything.

---

### STEP 5 — Screen modifications

After the base implementation, apply these specific changes:

[LIST YOUR MODIFICATIONS BELOW — copy and fill in as many as needed]

| Screen Name       | Change Type         | Instructions                              |
|-------------------|---------------------|-------------------------------------------|
| [Screen 1 name]   | Layout change       | [e.g. Move the CTA button above the fold] |
| [Screen 2 name]   | New component       | [e.g. Add a notification badge to avatar] |
| [Screen 3 name]   | Copy change         | [e.g. Rename "Submit" → "Send Request"]   |
| [Screen 4 name]   | Remove element      | [e.g. Remove the sidebar on mobile]       |
| [Screen 5 name]   | Add interaction     | [e.g. Add a dropdown on clicking the ⋯ menu] |

---

### STEP 6 — Quality checklist

After all screens are implemented, verify:

- [ ] All routes are reachable and linked correctly
- [ ] No console errors on any page
- [ ] Responsive: works at 375px, 768px, 1280px
- [ ] All interactive elements have hover/focus/active states
- [ ] Color contrast meets WCAG AA (4.5:1 for text)
- [ ] All images/icons have alt text
- [ ] Fonts load correctly (add fallback stack)
- [ ] No hardcoded pixel values that break on zoom

Report any failures and fix them.

---

### CONTEXT about my project

- App name: [YOUR APP NAME]
- App purpose: [1–2 sentences — what does this app do?]
- Target users: [e.g. operations managers, students, customers]
- Tech stack already decided: [e.g. React 18, Vite, Tailwind CSS 3]
- Backend / API: [e.g. REST API at /api/v1, Firebase, none yet]
- Auth: [e.g. JWT, Google OAuth, none]
- Special constraints: [e.g. must support RTL, dark mode required,
  must work offline, IE11 support, etc.]

---

### HOW TO HANDLE AMBIGUITY

- If a screen has interactive states not visible in the static design
  (hover, error, loading, empty), implement sensible defaults and add
  a TODO comment: `// TODO: confirm [X] state design with Stitch`
- If two screens look identical, ask me before de-duplicating them
- If a Stitch component maps to an existing library component
  (e.g. MUI, shadcn/ui, Radix), prefer the library version unless
  I say otherwise
- Always prefer composition over duplication

---

BEGIN with Step 1. Fetch the Stitch screens and output the inventory table.
```

---

## HOW TO USE THIS PROMPT

1. **Open Cursor** in your project folder
2. Switch to **Agent mode** (Cmd/Ctrl + Shift + P → "Agent")
3. Fill in every `[ ]` placeholder above
4. Paste the filled prompt into the chat
5. Let the agent complete Step 1 (inventory), review it, then continue

## TIPS FOR EACH STEP

| Step | What to review before saying "continue" |
|------|------------------------------------------|
| 1 | Confirm every screen is listed; note any missing ones |
| 2 | Check the token names match your mental model |
| 3 | Approve/adjust folder structure & library choices |
| 4 | Spot-check 2–3 components against the Stitch designs |
| 5 | Verify each modification was applied correctly |
| 6 | Run the checklist yourself in the browser |

## COMMON FOLLOW-UP PROMPTS

After the initial run, use these to refine:

```
# Add a new screen
Fetch the Stitch screen named "[SCREEN NAME]" and implement it
following the same design system and code patterns used for the
existing screens.
```

```
# Fix a specific screen
The [SCREEN NAME] page doesn't match the Stitch design.
Re-fetch it from Stitch and reconcile the implementation.
Pay special attention to: [spacing / colors / component used].
```

```
# Update the design system
The Stitch project has been updated with new brand colors.
Re-fetch the design tokens and propagate the changes to
tokens.css and all components that use the old values.
```

```
# Responsive pass
Review all pages and ensure they are fully responsive.
Use the breakpoints defined in tokens.css.
Mobile-first approach. No horizontal scroll below 375px.
```
