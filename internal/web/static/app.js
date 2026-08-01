"use strict";

const $ = (s) => document.querySelector(s);
const $$ = (s) => [...document.querySelectorAll(s)];
const api = async (url, opts) => {
  const r = await fetch(url, opts);
  const j = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(j.error || r.statusText);
  return j;
};
const postJSON = (url, body) =>
  api(url, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) });

function toast(msg) {
  const t = $("#toast");
  t.textContent = msg;
  t.classList.remove("hidden");
  clearTimeout(toast._t);
  toast._t = setTimeout(() => t.classList.add("hidden"), 4000);
}

const CATS = ["currency", "potion", "food", "material", "gear", "character", "misc"];
const catClass = (c) => "cat-" + (CATS.includes(c) ? c : "misc");
const escapeHtml = (s) =>
  String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));

// ── View & section switching ─────────────────────────────────────────────
let dbLoaded = false;

function initTabs() {
  $$(".tab").forEach((t) => (t.onclick = () => showView(t.dataset.view)));
  $$(".section-link").forEach((l) => (l.onclick = () => showSection(l.dataset.section)));
}
function showView(v) {
  $$(".tab").forEach((t) => t.classList.toggle("active", t.dataset.view === v));
  $("#view-editor").classList.toggle("hidden", v !== "editor");
  $("#view-database").classList.toggle("hidden", v !== "database");
  if (v === "database" && !dbLoaded) {
    loadTables().catch((e) => toast("Load failed: " + e.message));
    dbLoaded = true;
  }
}
function showSection(s) {
  $$(".section-link").forEach((l) => l.classList.toggle("active", l.dataset.section === s));
  $$("#view-editor .panel").forEach((p) => p.classList.toggle("hidden", p.id !== "panel-" + s));
}

// ── Shared editor row: name + quantity stepper + optional label edit ───────
function stepperRow({ cat, name, cid, known, value, onCommit, onLabel }) {
  const row = document.createElement("div");
  row.className = "row";

  const dot = document.createElement("span");
  dot.className = "cat-dot " + catClass(cat);
  row.appendChild(dot);

  const names = document.createElement("div");
  names.className = "names";
  const iname = document.createElement("span");
  iname.className = "iname" + (known ? "" : " unknown");
  iname.textContent = name;
  if (onLabel) {
    const edit = document.createElement("button");
    edit.className = "edit-label";
    edit.textContent = "✎ name";
    edit.onclick = () => {
      const v = prompt(`Name for CID ${cid}`, known ? name : "");
      if (v !== null) onLabel(v.trim());
    };
    iname.appendChild(edit);
  }
  const icid = document.createElement("span");
  icid.className = "icid";
  icid.textContent = "CID " + cid;
  names.appendChild(iname);
  names.appendChild(icid);
  row.appendChild(names);

  const qty = document.createElement("div");
  qty.className = "qty";
  const input = document.createElement("input");
  input.type = "number";
  input.value = value;
  const commit = async () => {
    const v = parseInt(input.value, 10);
    if (!Number.isFinite(v) || String(v) === String(value)) return;
    try {
      await onCommit(v);
      value = v;
      input.classList.remove("saved");
      void input.offsetWidth;
      input.classList.add("saved");
    } catch (e) {
      input.value = value;
      toast("Update failed: " + e.message);
    }
  };
  const bump = (d) => {
    input.value = Math.max(0, (parseInt(input.value, 10) || 0) + d);
    commit();
  };
  const minus = document.createElement("button");
  minus.textContent = "−";
  minus.onclick = () => bump(-1);
  const plus = document.createElement("button");
  plus.textContent = "+";
  plus.onclick = () => bump(1);
  input.onchange = commit;
  input.onkeydown = (e) => {
    if (e.key === "Enter") input.blur();
  };
  qty.appendChild(minus);
  qty.appendChild(input);
  qty.appendChild(plus);
  row.appendChild(qty);
  return row;
}

// ── Editor: Currency panel ────────────────────────────────────────────────
async function renderCurrency() {
  const el = $("#panel-currency");
  el.innerHTML = `<h2>Currency</h2><p class="panel-sub">Account currencies (gold, tokens…). Edit an amount and Enter, then Write to save file.</p>`;
  const { currencies } = await api("/api/game/currency");
  const rows = document.createElement("div");
  rows.className = "rows";
  for (const c of currencies || []) {
    rows.appendChild(
      stepperRow({
        cat: c.category, name: c.name, cid: c.cid, known: c.known, value: c.amount,
        onCommit: (v) => postJSON("/api/game/currency", { cid: c.cid, amount: v }),
        onLabel: (name) => postJSON("/api/game/label", { cid: c.cid, name }).then(renderCurrency),
      })
    );
  }
  el.appendChild(rows);
}

// ── Editor: Consumables panel ─────────────────────────────────────────────
async function renderConsumables() {
  const el = $("#panel-consumables");
  el.innerHTML = `<h2>Consumables</h2><p class="panel-sub">Potions, cooked food and materials. Grouped by category; names unknown until seeded or labelled (✎).</p>`;
  const { items } = await api("/api/game/consumables");
  const groups = {};
  for (const it of items || []) (groups[it.category] ||= []).push(it);
  const order = ["food", "potion", "material", "misc"];
  for (const cat of order) {
    if (!groups[cat]) continue;
    const title = document.createElement("div");
    title.className = "group-title";
    title.textContent = cat + " (" + groups[cat].length + ")";
    el.appendChild(title);
    const rows = document.createElement("div");
    rows.className = "rows";
    for (const it of groups[cat]) {
      rows.appendChild(
        stepperRow({
          cat: it.category, name: it.name, cid: it.cid, known: it.known, value: it.count,
          onCommit: (v) => postJSON("/api/game/stack", { kind: it.kind, id: it.id, count: v }),
          onLabel: (name) => postJSON("/api/game/label", { cid: it.cid, name }).then(renderConsumables),
        })
      );
    }
    el.appendChild(rows);
  }
}

// ── Editor: Characters panel (read-only) ──────────────────────────────────
async function renderCharacters() {
  const el = $("#panel-characters");
  el.innerHTML = `<h2>Characters</h2><p class="panel-sub">Read-only — owned characters and their progression.</p>`;
  const { characters } = await api("/api/game/characters");
  const grid = document.createElement("div");
  grid.className = "cards";
  for (const c of characters || []) {
    const card = document.createElement("div");
    card.className = "card";
    const stat = (label, v) => `<div><dt>${label}</dt><dd>${v}</dd></div>`;
    card.innerHTML =
      `<div class="card-head"><span class="cat-dot ${catClass(c.category)}"></span>` +
      `<span class="iname ${c.known ? "" : "unknown"}">${escapeHtml(c.name)}</span></div>` +
      `<div class="icid">CID ${c.cid}</div>` +
      `<dl class="stats">${stat("Level", c.level)}${stat("EXP", c.exp)}${stat("HP", c.hp)}` +
      `${stat("Ascend", c.ascend)}${stat("Transcend", c.transcend)}${stat("Soldier", c.soldierGrade)}</dl>`;
    grid.appendChild(card);
  }
  el.appendChild(grid);
}

// ── Editor: Team panel (read-only) ────────────────────────────────────────
async function renderTeam() {
  const el = $("#panel-team");
  el.innerHTML = `<h2>Team</h2><p class="panel-sub">Read-only — saved team pages (three slots each).</p>`;
  const { teams } = await api("/api/game/teams");
  for (const p of teams || []) {
    const title = document.createElement("div");
    title.className = "group-title";
    title.textContent = "Page " + p.pageId;
    el.appendChild(title);
    const slots = document.createElement("div");
    slots.className = "slots";
    for (const s of p.slots || []) {
      const slot = document.createElement("div");
      slot.className = "slot";
      if (s.empty) {
        slot.classList.add("empty");
        slot.textContent = "— empty —";
      } else {
        slot.innerHTML =
          `<span class="cat-dot ${catClass(s.category)}"></span>` +
          `<span class="iname ${s.known ? "" : "unknown"}">${escapeHtml(s.name)}</span>` +
          `<span class="lv">Lv ${s.level}</span>`;
      }
      slots.appendChild(slot);
    }
    el.appendChild(slots);
  }
}

async function loadEditor() {
  await Promise.all([renderCurrency(), renderConsumables(), renderCharacters(), renderTeam()]);
}

// ── Database: generic table browser ───────────────────────────────────────
const state = { table: null, limit: 200, offset: 0, total: 0, columns: [], pkCols: [], tables: [] };

async function loadInfo() {
  const info = await api("/api/info");
  $("#path").textContent = info.path;
  $("#path").title = info.path;
  state.tables = info.tables || [];
}
async function loadTables() {
  if (!state.tables.length) await loadInfo();
  renderTableList();
}
function renderTableList() {
  const ul = $("#tables");
  const filter = $("#filter").value.toLowerCase();
  ul.innerHTML = "";
  for (const name of state.tables) {
    if (filter && !name.toLowerCase().includes(filter)) continue;
    const li = document.createElement("li");
    li.textContent = name;
    li.title = name;
    if (name === state.table) li.classList.add("active");
    li.onclick = () => openTable(name);
    ul.appendChild(li);
  }
}
async function openTable(name) {
  state.table = name;
  state.offset = 0;
  renderTableList();
  await loadPage();
}
async function loadPage() {
  const { table, limit, offset } = state;
  const data = await api(`/api/table?name=${encodeURIComponent(table)}&limit=${limit}&offset=${offset}`);
  state.columns = data.columns;
  state.total = data.total;
  state.pkCols = data.columns.filter((c) => c.primaryKey).map((c) => c.name);
  $("#table-name").textContent = table;
  $("#hint").classList.add("hidden");
  renderGrid(data.columns, data.rows);
  renderPager();
}
function renderPager() {
  const p = $("#pager");
  p.classList.remove("hidden");
  const from = state.total ? state.offset + 1 : 0;
  const to = Math.min(state.offset + state.limit, state.total);
  $("#range").textContent = `${from}–${to} of ${state.total}`;
  $("#prev").disabled = state.offset <= 0;
  $("#next").disabled = to >= state.total;
}
function renderGrid(columns, rows) {
  const grid = $("#grid");
  grid.innerHTML = "";
  const thead = document.createElement("thead");
  const htr = document.createElement("tr");
  for (const c of columns) {
    const th = document.createElement("th");
    th.textContent = c.name;
    if (c.primaryKey) {
      const s = document.createElement("span");
      s.className = "pk";
      s.textContent = "KEY";
      th.appendChild(s);
    }
    const t = document.createElement("span");
    t.className = "type";
    t.textContent = c.type || "";
    th.appendChild(t);
    htr.appendChild(th);
  }
  thead.appendChild(htr);
  grid.appendChild(thead);

  const tbody = document.createElement("tbody");
  for (const row of rows || []) {
    const tr = document.createElement("tr");
    const pk = {};
    columns.forEach((c, i) => {
      if (c.primaryKey) pk[c.name] = row[i];
    });
    columns.forEach((c, i) => {
      const td = document.createElement("td");
      setCellText(td, row[i]);
      if (c.primaryKey) {
        td.classList.add("pk");
      } else {
        td.classList.add("editable");
        td.title = "double-click to edit";
        td.ondblclick = () => beginEdit(td, c, pk);
      }
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  }
  grid.appendChild(tbody);
}
function setCellText(td, val) {
  if (val === null || val === undefined) {
    td.textContent = "NULL";
    td.classList.add("null");
  } else {
    td.textContent = String(val);
    td.classList.remove("null");
  }
}
function beginEdit(td, col, pk) {
  const original = td.classList.contains("null") ? "" : td.textContent;
  td.contentEditable = "true";
  td.focus();
  const range = document.createRange();
  range.selectNodeContents(td);
  const sel = window.getSelection();
  sel.removeAllRanges();
  sel.addRange(range);

  const finish = async (commit) => {
    td.contentEditable = "false";
    td.removeEventListener("keydown", onKey);
    td.removeEventListener("blur", onBlur);
    const text = td.textContent;
    if (!commit || text === original) {
      setCellText(td, original === "" && td.classList.contains("null") ? null : original);
      return;
    }
    try {
      let value = text;
      if (/^-?\d+$/.test(text)) value = parseInt(text, 10);
      else if (/^-?\d*\.\d+$/.test(text)) value = parseFloat(text);
      await postJSON("/api/update", { table: state.table, pk, column: col.name, value });
      setCellText(td, value);
      td.classList.remove("saved");
      void td.offsetWidth;
      td.classList.add("saved");
    } catch (e) {
      setCellText(td, original);
      toast("Update failed: " + e.message);
    }
  };
  const onKey = (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      finish(true);
    } else if (e.key === "Escape") {
      e.preventDefault();
      finish(false);
    }
  };
  const onBlur = () => finish(true);
  td.addEventListener("keydown", onKey);
  td.addEventListener("blur", onBlur);
}

// ── Wiring ────────────────────────────────────────────────────────────────
$("#filter").addEventListener("input", renderTableList);
$("#prev").onclick = () => {
  state.offset = Math.max(0, state.offset - state.limit);
  loadPage();
};
$("#next").onclick = () => {
  state.offset += state.limit;
  loadPage();
};
$("#save-btn").onclick = async () => {
  const s = $("#save-status");
  s.textContent = "Writing…";
  s.className = "status";
  try {
    await postJSON("/api/save", {});
    s.textContent = "Saved ✓";
    s.className = "status ok";
  } catch (e) {
    s.textContent = "Failed";
    s.className = "status err";
    toast("Save failed: " + e.message);
  }
  setTimeout(() => (s.textContent = ""), 4000);
};

initTabs();
loadInfo().catch((e) => toast("Load failed: " + e.message));
loadEditor().catch((e) => toast("Editor load failed: " + e.message));
