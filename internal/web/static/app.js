"use strict";

const $ = (s) => document.querySelector(s);
const api = async (url, opts) => {
  const r = await fetch(url, opts);
  const j = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(j.error || r.statusText);
  return j;
};

const state = {
  table: null,
  limit: 200,
  offset: 0,
  total: 0,
  columns: [],
  pkCols: [],
};

function toast(msg) {
  const t = $("#toast");
  t.textContent = msg;
  t.classList.remove("hidden");
  clearTimeout(toast._t);
  toast._t = setTimeout(() => t.classList.add("hidden"), 4000);
}

async function loadInfo() {
  const info = await api("/api/info");
  $("#path").textContent = info.path;
  $("#path").title = info.path;
  renderTableList(info.tables || []);
}

function renderTableList(tables) {
  const ul = $("#tables");
  const filter = $("#filter").value.toLowerCase();
  ul.innerHTML = "";
  for (const name of tables) {
    if (filter && !name.toLowerCase().includes(filter)) continue;
    const li = document.createElement("li");
    li.textContent = name;
    li.title = name;
    if (name === state.table) li.classList.add("active");
    li.onclick = () => openTable(name);
    ul.appendChild(li);
  }
  renderTableList._all = tables;
}

async function openTable(name) {
  state.table = name;
  state.offset = 0;
  renderTableList(renderTableList._all || []);
  await loadPage();
}

async function loadPage() {
  const { table, limit, offset } = state;
  const data = await api(
    `/api/table?name=${encodeURIComponent(table)}&limit=${limit}&offset=${offset}`
  );
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
      s.className = "pk"; s.textContent = "KEY"; th.appendChild(s);
    }
    const t = document.createElement("span");
    t.className = "type"; t.textContent = c.type || ""; th.appendChild(t);
    htr.appendChild(th);
  }
  thead.appendChild(htr);
  grid.appendChild(thead);

  const tbody = document.createElement("tbody");
  for (const row of rows || []) {
    const tr = document.createElement("tr");
    const pk = {};
    columns.forEach((c, i) => { if (c.primaryKey) pk[c.name] = row[i]; });
    columns.forEach((c, i) => {
      const td = document.createElement("td");
      const val = row[i];
      setCellText(td, val);
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
    if (!commit || text === original) { setCellText(td, original === "" && td.classList.contains("null") ? null : original); return; }
    try {
      let value = text;
      if (/^-?\d+$/.test(text)) value = parseInt(text, 10);
      else if (/^-?\d*\.\d+$/.test(text)) value = parseFloat(text);
      await api("/api/update", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ table: state.table, pk, column: col.name, value }),
      });
      setCellText(td, value);
      td.classList.remove("saved"); void td.offsetWidth; td.classList.add("saved");
    } catch (e) {
      setCellText(td, original);
      toast("Update failed: " + e.message);
    }
  };
  const onKey = (e) => {
    if (e.key === "Enter") { e.preventDefault(); finish(true); }
    else if (e.key === "Escape") { e.preventDefault(); finish(false); }
  };
  const onBlur = () => finish(true);
  td.addEventListener("keydown", onKey);
  td.addEventListener("blur", onBlur);
}

$("#filter").addEventListener("input", () => renderTableList(renderTableList._all || []));
$("#prev").onclick = () => { state.offset = Math.max(0, state.offset - state.limit); loadPage(); };
$("#next").onclick = () => { state.offset += state.limit; loadPage(); };
$("#save-btn").onclick = async () => {
  const s = $("#save-status");
  s.textContent = "Writing…"; s.className = "status";
  try {
    await api("/api/save", { method: "POST" });
    s.textContent = "Saved ✓"; s.className = "status ok";
  } catch (e) {
    s.textContent = "Failed"; s.className = "status err";
    toast("Save failed: " + e.message);
  }
  setTimeout(() => { s.textContent = ""; }, 4000);
};

loadInfo().catch((e) => toast("Load failed: " + e.message));
