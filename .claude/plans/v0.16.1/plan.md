# v0.16.1 — fix: writing the save reveals the inventory screen

## Context

On any panel that is **not** the inventory (e.g. *Monnaies*, *Titres*), after
editing something and clicking **« Écrire dans la save »**, the *Inventaire*
screen appears rendered next to the current panel.

## Root cause

`setView(v)` hides every inactive screen by toggling the `hidden` class
(`app.js:182`). Every `RENDER.*` function rebuilds its screen with
`el.className = "screen …"`, which **clobbers** that `hidden` class. This is
harmless because `RENDER` is only ever called for the *current* view — with one
exception:

The save handler (`app.js:1416`) re-renders `#screen-inv` unconditionally, so
the inventory baseline (amber "modified" dots) is refreshed after the write:

```js
invBaseline = {}; // written to disk → new baseline
if (loaded.has("inv")) RENDER.inv().catch(() => {});
```

When the current view is *not* inv, this re-render strips `hidden` from
`#screen-inv`, revealing it beside the active panel.

## Objective

Refresh the inventory baseline without un-hiding the inventory screen.

## Scope

- **In**: the `#save-btn` click handler in `internal/web/static/app.js`.
- **Out**: no change to the `RENDER.*` className pattern, no backend change.

## Layer

`web`.

## Fix

Only re-render inv immediately when it is the active view; otherwise drop it
from `loaded` so it re-renders (and recaptures its baseline) the next time it is
shown.

```js
invBaseline = {}; // written to disk → new baseline
if (currentView === "inv") RENDER.inv().catch(() => {});
else loaded.delete("inv"); // recapture baseline on next view
```

## Verification

- `just ci` (Go gates — no Go change, must still pass).
- Manual: on *Monnaies* and *Titres*, edit → « Écrire dans la save » → only the
  current panel stays visible; the inventory does not appear.
- Manual: on *Inventaire*, edit → save → amber dots clear (baseline refreshed).
- Manual: after saving from *Titres*, navigate to *Inventaire* → amber dots are
  clean (baseline was recaptured).

## Commit

`fix(web): don't reveal the inventory screen when writing the save`
