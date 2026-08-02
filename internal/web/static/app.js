"use strict";

const $ = (s) => document.querySelector(s);
const $$ = (s) => [...document.querySelectorAll(s)];

const api = async (url, opts) => {
  const r = await fetch(url, opts);
  const j = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(j.error || r.statusText);
  return j;
};
// postJSON also drives the "N modif." badge: game/table mutations bump it,
// writing to disk resets it. Session calls (config/open/save) are excluded.
const MUTATING = (url) => url.startsWith("/api/game/") || url === "/api/update";
async function postJSON(url, body) {
  const j = await api(url, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) });
  if (MUTATING(url)) bumpModif(1);
  return j;
}

function toast(msg) {
  const t = $("#toast");
  t.textContent = msg;
  t.classList.remove("hidden");
  clearTimeout(toast._t);
  toast._t = setTimeout(() => t.classList.add("hidden"), 4000);
}

const escapeHtml = (s) =>
  String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));

let lang = localStorage.getItem("dsa-lang") || "fr";
const displayName = (it) =>
  (lang === "fr" ? it.nameFr : it.nameEn) || it.nameEn || it.nameFr || "CID " + it.cid;

// ── Modif badge (derived, client-side; reset on write) ─────────────────────
let modifCount = 0;
function bumpModif(n) {
  modifCount += n;
  const b = $("#modif-badge");
  b.textContent = modifCount + " modif.";
  b.classList.toggle("hidden", modifCount === 0);
}
function resetModif() { modifCount = 0; $("#modif-badge").classList.add("hidden"); }

// ── Icons (sprite atlas) ───────────────────────────────────────────────────
let iconSize = 64;
const ICON_DISP = 40;
const iconStyle = (it, disp = ICON_DISP) =>
  `object-fit:none;object-position:${it.iconX}px ${it.iconY}px;` +
  `width:${iconSize}px;height:${iconSize}px;zoom:${disp / iconSize}`;
function iconEl(it, disp) {
  if (!it.icon) {
    const d = document.createElement("span");
    d.className = "cat-dot";
    return d;
  }
  const img = document.createElement("img");
  img.className = "ic";
  img.src = "/sprite.webp";
  img.alt = "";
  img.style.cssText = iconStyle(it, disp);
  return img;
}
const iconHTML = (it, disp) =>
  it.icon ? `<img class="ic" src="/sprite.webp" alt="" style="${iconStyle(it, disp)}">` : `<span class="cat-dot"></span>`;

// ── Session state & view routing ───────────────────────────────────────────
let saveOpen = false;
let currentView = "home";
const RENDER = {}; // view -> async render fn (registered below)
const loaded = new Set();

function setView(v) {
  if (v !== "home" && !saveOpen) return;
  currentView = v;
  $$(".tab").forEach((t) => t.classList.toggle("active", t.dataset.view === v));
  $$(".screen").forEach((s) => s.classList.toggle("hidden", s.id !== "screen-" + v));
  if (v !== "home" && RENDER[v] && !loaded.has(v)) {
    loaded.add(v);
    RENDER[v]().catch((e) => toast("Chargement échoué : " + e.message));
  }
}

function onSaveOpened(path) {
  saveOpen = true;
  $("#path").textContent = path;
  $("#path").title = path;
  $("#save-btn").classList.remove("hidden");
  document.body.classList.add("has-save");
  loaded.clear();
  resetModif();
  loadInfo().catch(() => {});
}

// Re-render everything after a language switch.
function reloadAll() {
  const keep = currentView;
  loaded.clear();
  catalogCache = consCatsCache = null;
  if (keep !== "home" && RENDER[keep]) {
    loaded.add(keep);
    RENDER[keep]().catch((e) => toast("Rechargement échoué : " + e.message));
  }
}

// ── Accueil : game folder + save picker ────────────────────────────────────
async function renderHome() {
  const el = $("#screen-home");
  el.className = "screen home";
  el.innerHTML = "";

  const head = document.createElement("div");
  head.className = "home-head";
  head.innerHTML = `<span class="overline">Éditeur de sauvegarde</span><h1>Choisis ta sauvegarde</h1>`;
  el.appendChild(head);

  const cfg = await api("/api/config");
  if (cfg.saveOpen && !saveOpen) onSaveOpened(cfg.savePath || "");

  // Folder box.
  const box = document.createElement("div");
  box.className = "folder-box";
  if (!cfg.gameDir) {
    box.innerHTML = `<div class="label">Dossier du jeu</div>
      <p class="sub">Indique le dossier d'installation de DragonSword (celui qui contient <span class="mono">DS/Content/Paks</span> et <span class="mono">DS/Saved/SaveGames</span>).</p>`;
  } else {
    box.innerHTML = `<div class="label">Dossier du jeu</div>
      <div class="cur-folder">${escapeHtml(cfg.gameDir)}</div>`;
  }
  const row = document.createElement("div");
  row.className = "folder-row";
  const input = document.createElement("input");
  input.placeholder = "/chemin/vers/DragonSword Awakening";
  input.value = cfg.gameDir || "";
  const btn = document.createElement("button");
  btn.className = "cta";
  btn.textContent = cfg.gameDir ? "Changer" : "Valider";
  btn.onclick = async () => {
    const dir = input.value.trim();
    if (!dir) return;
    btn.disabled = true;
    try {
      await postJSON("/api/config/game-dir", { dir });
      await renderHome();
    } catch (e) {
      toast("Dossier invalide : " + e.message);
    } finally {
      btn.disabled = false;
    }
  };
  input.onkeydown = (e) => { if (e.key === "Enter") btn.click(); };
  row.append(input, btn);
  box.appendChild(row);
  el.appendChild(box);

  // Slots.
  if (cfg.gameDir) {
    try {
      const { slots } = await api("/api/saves");
      if (!slots || !slots.length) {
        const p = document.createElement("p");
        p.className = "empty-note";
        p.textContent = "Aucune sauvegarde trouvée sous ce dossier.";
        el.appendChild(p);
      } else {
        const grid = document.createElement("div");
        grid.className = "slots";
        for (const s of slots) grid.appendChild(slotCard(s));
        el.appendChild(grid);
      }
    } catch (e) {
      const p = document.createElement("p");
      p.className = "empty-note";
      p.textContent = "Impossible de lister les sauvegardes : " + e.message;
      el.appendChild(p);
    }
  }

  // Warning banner.
  const warn = document.createElement("div");
  warn.className = "warn-banner";
  warn.innerHTML = `<span class="dot">!</span><p><b>Ferme le jeu avant d'écrire.</b> Le jeu garde la base ouverte pendant qu'il tourne. Au premier write, un backup horodaté <span class="mono">.bak</span> de l'original est créé automatiquement.</p>`;
  el.appendChild(warn);
}

function slotCard(s) {
  const card = document.createElement("div");
  card.className = "slot-card";
  const when = s.modTime ? new Date(s.modTime).toLocaleString() : "";
  const thumb = s.screenshot
    ? `<img class="slot-thumb" src="/api/screenshot?path=${encodeURIComponent(s.screenshot)}" alt="">`
    : `<div class="slot-thumb"></div>`;
  card.innerHTML =
    `<div class="slot-top"><span class="slot-n">SLOT ${s.slot}</span><span class="pill">compte ${escapeHtml(s.accountId)}</span></div>` +
    `<div class="slot-main">${thumb}<div class="slot-meta">` +
    `<span class="slot-name">Slot ${s.slot}</span><span class="slot-when">${escapeHtml(when)}</span></div></div>` +
    `<div class="slot-path" title="${escapeHtml(s.path)}">${escapeHtml(s.path)}</div>`;
  card.onclick = async () => {
    try {
      await api("/api/open", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ path: s.path }) });
      onSaveOpened(s.path);
      setView("inv");
    } catch (e) {
      toast("Ouverture échouée : " + e.message);
    }
  };
  return card;
}

// ── Shared stepper row ─────────────────────────────────────────────────────
function stepperRow({ item, name, cid, known, value, step = 1, onCommit, onLabel }) {
  const row = document.createElement("div");
  row.className = "row";
  row.appendChild(iconEl(item, 32));

  const names = document.createElement("div");
  names.className = "names";
  const iname = document.createElement("span");
  iname.className = "iname" + (known ? "" : " unknown");
  iname.textContent = name;
  if (onLabel) {
    const edit = document.createElement("button");
    edit.className = "edit-label";
    edit.textContent = "✎";
    edit.onclick = () => {
      const v = prompt(`Nom pour CID ${cid}`, known ? name : "");
      if (v !== null) onLabel(v.trim());
    };
    iname.appendChild(edit);
  }
  const icid = document.createElement("span");
  icid.className = "icid";
  icid.textContent = "CID " + cid;
  names.append(iname, icid);
  row.appendChild(names);
  row.appendChild(stepper(value, step, onCommit));
  return row;
}

function stepper(value, step, onCommit) {
  const qty = document.createElement("div");
  qty.className = "stepper";
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
      toast("Échec : " + e.message);
    }
  };
  const bump = (d) => { input.value = Math.max(0, (parseInt(input.value, 10) || 0) + d); commit(); };
  const minus = document.createElement("button");
  minus.textContent = "−";
  minus.onclick = () => bump(-step);
  const plus = document.createElement("button");
  plus.className = "plus";
  plus.textContent = "+";
  plus.onclick = () => bump(step);
  input.onchange = commit;
  input.onkeydown = (e) => { if (e.key === "Enter") input.blur(); };
  qty.append(minus, input, plus);
  return qty;
}

// ── Monnaies ───────────────────────────────────────────────────────────────
RENDER.money = async function () {
  const el = $("#screen-money");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Compte</span><h2>Monnaies</h2></div>`;
  const { currencies } = await api("/api/game/currency");
  const cards = document.createElement("div");
  cards.className = "cards";
  for (const c of currencies || []) {
    const card = document.createElement("div");
    card.className = "card";
    const head = document.createElement("div");
    head.className = "card-head";
    head.appendChild(iconEl(c, 40));
    const names = document.createElement("div");
    names.className = "names";
    names.innerHTML = `<span class="iname">${escapeHtml(displayName(c))}</span><span class="icid">CID ${c.cid}</span>`;
    head.appendChild(names);
    card.appendChild(head);
    // Gold steps by 10k, crystals/tokens by 50.
    const step = /gold|or\b/i.test(displayName(c)) ? 10000 : 50;
    card.appendChild(stepper(c.amount, step, (v) => postJSON("/api/game/currency", { cid: c.cid, amount: v })));
    cards.appendChild(card);
  }
  el.appendChild(cards);
};

// ── Inventaire (categories rail + list) ────────────────────────────────────
let catalogCache = null;
async function catalog() {
  if (!catalogCache) catalogCache = (await api("/api/game/catalog")).items || [];
  return catalogCache;
}
let consCatsCache = null;
async function consCategories() {
  if (!consCatsCache) consCatsCache = (await api("/api/game/consumable-categories")).categories || [];
  return consCatsCache;
}
let selectedCat = null;
const catLabel = (c) => (lang === "fr" ? c.labelFr : c.labelEn) || c.labelEn || c.key;

async function buildConsumableModel() {
  const [{ items: owned }, cat, cats] = await Promise.all([api("/api/game/consumables"), catalog(), consCategories()]);
  const catByCid = {};
  for (const it of cat) catByCid[it.cid] = it;
  const entries = {};
  const push = (key, e) => (entries[key] ||= []).push(e);
  const ownedStack = {};
  for (const it of owned || []) {
    if (it.kind === "cook") {
      push(it.group || "unsorted", { item: it, name: displayName(it), cid: it.cid, known: it.known, value: it.count, owned: it.count > 0, stackable: false,
        commit: (v) => postJSON("/api/game/stack", { kind: it.kind, id: it.id, count: v }) });
    } else {
      ownedStack[it.cid] = it.count;
      if (!catByCid[it.cid]) {
        push(it.group || "unsorted", { item: it, name: displayName(it), cid: it.cid, known: it.known, value: it.count, owned: it.count > 0, stackable: true,
          commit: (v) => postJSON("/api/game/stackable", { cid: it.cid, count: v }) });
      }
    }
  }
  for (const it of cat) {
    if (!it.group) continue;
    const c = ownedStack[it.cid] || 0;
    push(it.group, { item: it, name: displayName(it), cid: it.cid, known: it.known, value: c, owned: c > 0, stackable: true,
      commit: (v) => postJSON("/api/game/stackable", { cid: it.cid, count: v }) });
  }
  for (const k in entries) entries[k].sort((a, b) => (b.owned - a.owned) || a.name.localeCompare(b.name));
  return { entries, cats };
}

RENDER.inv = async function () {
  const el = $("#screen-inv");
  el.className = "screen inv";
  el.innerHTML = "";
  const model = await buildConsumableModel();
  const cats = model.cats.filter((c) => (model.entries[c.key] || []).length);

  const rail = document.createElement("aside");
  rail.className = "inv-rail";
  const main = document.createElement("div");
  main.className = "inv-main";
  el.append(rail, main);

  if (!cats.length) { main.innerHTML = `<p class="empty-note">Aucun consommable.</p>`; return; }
  if (!selectedCat || !cats.some((c) => c.key === selectedCat)) selectedCat = cats[0].key;

  for (const c of cats) {
    const list = model.entries[c.key];
    const ownedN = list.filter((e) => e.owned).length;
    const b = document.createElement("button");
    b.className = "cat-link" + (c.key === selectedCat ? " active" : "");
    b.dataset.key = c.key;
    b.innerHTML = `<span title="${escapeHtml(catLabel(c))}">${escapeHtml(catLabel(c))}</span><span class="n">${ownedN}/${list.length}</span>`;
    b.onclick = () => {
      selectedCat = c.key;
      $$(".cat-link").forEach((x) => x.classList.toggle("active", x.dataset.key === c.key));
      fillInvMain(main, model, c);
    };
    rail.appendChild(b);
  }
  fillInvMain(main, model, cats.find((c) => c.key === selectedCat));
};

function fillInvMain(main, model, cat) {
  const list = model.entries[cat.key] || [];
  const ownedN = list.filter((e) => e.owned).length;
  main.innerHTML = "";
  const bar = document.createElement("div");
  bar.className = "inv-toolbar";
  bar.innerHTML = `<h2>${escapeHtml(catLabel(cat))}</h2><span class="count">${ownedN} possédés / ${list.length}</span>`;
  const fillWrap = document.createElement("div");
  fillWrap.className = "inv-fill";
  const fillN = document.createElement("input");
  fillN.type = "number"; fillN.value = 999; fillN.min = 0;
  const fillBtn = document.createElement("button");
  fillBtn.textContent = "Remplir";
  fillBtn.onclick = async () => {
    const n = Math.max(0, parseInt(fillN.value, 10) || 0);
    const stacks = list.filter((e) => e.stackable);
    if (!stacks.length) return toast("Rien à remplir ici.");
    if (!confirm(`Mettre les ${stacks.length} objets empilables de « ${catLabel(cat)} » à ${n} ?`)) return;
    try { for (const e of stacks) await postJSON("/api/game/stackable", { cid: e.cid, count: n }); await RENDER.inv(); }
    catch (err) { toast("Remplissage échoué : " + err.message); }
  };
  fillWrap.append(fillN, fillBtn);
  bar.appendChild(fillWrap);
  main.appendChild(bar);

  const rows = document.createElement("div");
  rows.className = "rows";
  rows.style.overflow = "auto";
  for (const e of list) {
    const row = stepperRow({ item: e.item, name: e.name, cid: e.cid, known: e.known, value: e.value,
      onCommit: async (v) => { await e.commit(v); e.value = v; e.owned = v > 0; row.classList.toggle("not-owned", !e.owned); },
      onLabel: (name) => postJSON("/api/game/label", { cid: e.cid, name }).then(() => RENDER.inv()) });
    if (!e.owned) row.classList.add("not-owned");
    rows.appendChild(row);
  }
  main.appendChild(rows);
}

// ── Personnages (read-only) ────────────────────────────────────────────────
RENDER.chars = async function () {
  const el = $("#screen-chars");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Roster</span><h2>Personnages</h2><span class="sub">Lecture seule — personnages possédés et progression.</span></div>`;
  const { characters } = await api("/api/game/characters");
  const grid = document.createElement("div");
  grid.className = "cards";
  for (const c of characters || []) {
    const card = document.createElement("div");
    card.className = "card";
    const stat = (l, v) => `<div><dt class="label">${l}</dt><dd class="mono">${v}</dd></div>`;
    card.innerHTML =
      `<div class="card-head">${iconHTML(c, 40)}<div class="names"><span class="iname ${c.known ? "" : "unknown"}">${escapeHtml(displayName(c))}</span><span class="icid">CID ${c.cid}</span></div></div>` +
      `<dl class="stats" style="display:grid;grid-template-columns:1fr 1fr;gap:8px;margin:0">${stat("Niveau", c.level)}${stat("EXP", c.exp)}${stat("PV", c.hp)}${stat("Ascension", c.ascend)}${stat("Transcend.", c.transcend)}${stat("Soldat", c.soldierGrade)}</dl>`;
    grid.appendChild(card);
  }
  el.appendChild(grid);
};

// ── Équipe (read-only) ─────────────────────────────────────────────────────
RENDER.team = async function () {
  const el = $("#screen-team");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Formation</span><h2>Équipe</h2><span class="sub">Lecture seule — pages d'équipe (trois slots chacune).</span></div>`;
  const { teams } = await api("/api/game/teams");
  for (const p of teams || []) {
    const title = document.createElement("div");
    title.className = "group-title";
    title.textContent = "Page " + p.pageId;
    el.appendChild(title);
    const rows = document.createElement("div");
    rows.className = "rows";
    for (const s of p.slots || []) {
      const slot = document.createElement("div");
      slot.className = "slot" + (s.empty ? " empty" : "");
      if (s.empty) slot.textContent = "— vide —";
      else slot.innerHTML = `${iconHTML(s, 32)}<span class="iname ${s.known ? "" : "unknown"}">${escapeHtml(displayName(s))}</span><span class="lv">Niv ${s.level}</span>`;
      rows.appendChild(slot);
    }
    el.appendChild(rows);
  }
};

// ── Équipement & Gemmes (editable) ─────────────────────────────────────────
function namesCell(name, cid, known, onLabel) {
  const names = document.createElement("div");
  names.className = "names";
  const iname = document.createElement("span");
  iname.className = "iname" + (known ? "" : " unknown");
  iname.textContent = name;
  if (onLabel) {
    const edit = document.createElement("button");
    edit.className = "edit-label";
    edit.textContent = "✎";
    edit.onclick = () => { const v = prompt(`Nom pour CID ${cid}`, known ? name : ""); if (v !== null) onLabel(v.trim()); };
    iname.appendChild(edit);
  }
  const icid = document.createElement("span");
  icid.className = "icid";
  icid.textContent = "CID " + cid;
  names.append(iname, icid);
  return names;
}
function eqNumField(label, value, onCommit) {
  const wrap = document.createElement("label");
  wrap.className = "eq-field";
  wrap.innerHTML = `<span>${label}</span>`;
  const input = document.createElement("input");
  input.type = "number"; input.value = value;
  const commit = async () => {
    const v = parseInt(input.value, 10);
    if (!Number.isFinite(v) || String(v) === String(value)) return;
    try { await onCommit(v); value = v; input.classList.remove("saved"); void input.offsetWidth; input.classList.add("saved"); }
    catch (e) { input.value = value; toast("Échec : " + e.message); }
  };
  input.onchange = commit;
  input.onkeydown = (e) => { if (e.key === "Enter") input.blur(); };
  wrap.appendChild(input);
  return wrap;
}
function lockToggle(checked, onChange) {
  const label = document.createElement("label");
  label.className = "lock";
  const cb = document.createElement("input");
  cb.type = "checkbox"; cb.checked = checked;
  cb.onchange = async () => { try { await onChange(cb.checked); } catch (e) { cb.checked = !cb.checked; toast("Échec : " + e.message); } };
  label.append(cb, document.createTextNode("verrou"));
  return label;
}
function statChips(cids) {
  const chips = document.createElement("div");
  chips.className = "stat-chips";
  for (const cid of cids) { if (!cid) continue; const c = document.createElement("span"); c.className = "chip"; c.title = "stat CID (lecture seule)"; c.textContent = cid; chips.appendChild(c); }
  return chips;
}

RENDER.equip = async function () {
  const el = $("#screen-equip");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Stuff</span><h2>Équipement</h2><span class="sub">Enchant, XP d'objet et verrou éditables ; les stats (puces) sont en lecture seule.</span></div>`;
  const { equipment } = await api("/api/game/equipment");
  const rows = document.createElement("div");
  rows.className = "rows";
  for (const e of equipment || []) {
    const row = document.createElement("div");
    row.className = "row";
    row.appendChild(iconEl(e, 32));
    row.appendChild(namesCell(displayName(e), e.cid, e.known, (name) => postJSON("/api/game/label", { cid: e.cid, name }).then(() => RENDER.equip())));
    const fields = document.createElement("div");
    fields.className = "eq-fields";
    fields.appendChild(eqNumField("Enchant", e.enchantLevel, (v) => postJSON("/api/game/equipment", { dbid: e.dbid, field: "enchant", value: v })));
    fields.appendChild(eqNumField("XP", e.exp, (v) => postJSON("/api/game/equipment", { dbid: e.dbid, field: "exp", value: v })));
    fields.appendChild(lockToggle(e.isLock, (locked) => postJSON("/api/game/equipment", { dbid: e.dbid, field: "lock", value: locked ? 1 : 0 })));
    row.appendChild(fields);
    row.appendChild(statChips([e.mainStatCid, ...(e.subStatCids || [])]));
    rows.appendChild(row);
  }
  el.appendChild(rows);
};

RENDER.gems = async function () {
  const el = $("#screen-gems");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Sertissage</span><h2>Gemmes</h2><span class="sub">Gemmes serties (<span class="mono">tb_gem</span>). Le verrou est éditable. Les gemmes d'inventaire sont sous <b>Inventaire</b>.</span></div>`;
  const { gems } = await api("/api/game/gems");
  if (!gems || !gems.length) { el.innerHTML += `<p class="empty-note">Aucune gemme sertie. Les gemmes d'inventaire sont éditables sous Inventaire.</p>`; return; }
  const rows = document.createElement("div");
  rows.className = "rows";
  for (const gm of gems) {
    const row = document.createElement("div");
    row.className = "row";
    row.appendChild(iconEl(gm, 32));
    row.appendChild(namesCell(displayName(gm), gm.cid, gm.known, (name) => postJSON("/api/game/label", { cid: gm.cid, name }).then(() => RENDER.gems())));
    const fields = document.createElement("div");
    fields.className = "eq-fields";
    fields.appendChild(lockToggle(gm.isLock, (locked) => postJSON("/api/game/gem", { dbid: gm.dbid, locked })));
    row.appendChild(fields);
    row.appendChild(statChips([gm.statInfoCid]));
    rows.appendChild(row);
  }
  el.appendChild(rows);
};

// ── Cuisine (recipe details land in Phase 3) ───────────────────────────────
RENDER.cook = async function () {
  const el = $("#screen-cook");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Fourneaux</span><h2>Livre de recettes</h2></div>`;
  const btn = document.createElement("button");
  btn.className = "action-btn";
  btn.textContent = "Tout débloquer";
  btn.onclick = async () => {
    if (!confirm("Marquer toutes les recettes normales comme connues ?")) return;
    btn.disabled = true;
    try { await postJSON("/api/game/recipes/unlock-all", {}); toast("Recettes débloquées — clique « Écrire dans la save »."); }
    catch (e) { toast("Échec : " + e.message); }
    finally { btn.disabled = false; }
  };
  const note = document.createElement("p");
  note.className = "empty-note";
  note.innerHTML = "Marque chaque recette normale (grillé / bouilli / tranché) comme connue. Les détails par recette (matériaux requis / possédés) arrivent en Phase 3.";
  el.append(btn, note);
};

// ── SQL brut (generic table browser) ───────────────────────────────────────
const sql = { table: null, limit: 200, offset: 0, total: 0, columns: [], pkCols: [], tables: [] };
async function loadInfo() {
  const info = await api("/api/info");
  iconSize = info.iconSize || 64;
  sql.tables = info.tables || [];
}
RENDER.sql = async function () {
  if (!sql.tables.length) await loadInfo();
  renderTableList();
};
function renderTableList() {
  const ul = $("#tables");
  const filter = $("#filter").value.toLowerCase();
  ul.innerHTML = "";
  for (const name of sql.tables) {
    if (filter && !name.toLowerCase().includes(filter)) continue;
    const li = document.createElement("li");
    li.textContent = name; li.title = name;
    if (name === sql.table) li.classList.add("active");
    li.onclick = () => openTable(name);
    ul.appendChild(li);
  }
}
async function openTable(name) { sql.table = name; sql.offset = 0; renderTableList(); await loadPage(); }
async function loadPage() {
  const { table, limit, offset } = sql;
  const data = await api(`/api/table?name=${encodeURIComponent(table)}&limit=${limit}&offset=${offset}`);
  sql.columns = data.columns; sql.total = data.total;
  sql.pkCols = data.columns.filter((c) => c.primaryKey).map((c) => c.name);
  $("#table-name").textContent = table;
  $("#hint").classList.add("hidden");
  renderGrid(data.columns, data.rows);
  renderPager();
}
function renderPager() {
  $("#pager").classList.remove("hidden");
  const from = sql.total ? sql.offset + 1 : 0;
  const to = Math.min(sql.offset + sql.limit, sql.total);
  $("#range").textContent = `${from}–${to} sur ${sql.total}`;
  $("#prev").disabled = sql.offset <= 0;
  $("#next").disabled = to >= sql.total;
}
function renderGrid(columns, rows) {
  const grid = $("#grid");
  grid.innerHTML = "";
  const thead = document.createElement("thead");
  const htr = document.createElement("tr");
  for (const c of columns) {
    const th = document.createElement("th");
    th.textContent = c.name;
    if (c.primaryKey) { const s = document.createElement("span"); s.className = "pk"; s.textContent = "KEY"; th.appendChild(s); }
    const t = document.createElement("span"); t.className = "type"; t.textContent = c.type || ""; th.appendChild(t);
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
      setCellText(td, row[i]);
      if (c.primaryKey) td.classList.add("pk");
      else { td.classList.add("editable"); td.title = "double-clic pour éditer"; td.ondblclick = () => beginEdit(td, c, pk); }
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  }
  grid.appendChild(tbody);
}
function setCellText(td, val) {
  if (val === null || val === undefined) { td.textContent = "NULL"; td.classList.add("null"); }
  else { td.textContent = String(val); td.classList.remove("null"); }
}
function beginEdit(td, col, pk) {
  const original = td.classList.contains("null") ? "" : td.textContent;
  td.contentEditable = "true"; td.focus();
  const range = document.createRange(); range.selectNodeContents(td);
  const s = window.getSelection(); s.removeAllRanges(); s.addRange(range);
  const finish = async (commit) => {
    td.contentEditable = "false";
    td.removeEventListener("keydown", onKey); td.removeEventListener("blur", onBlur);
    const text = td.textContent;
    if (!commit || text === original) { setCellText(td, original === "" && td.classList.contains("null") ? null : original); return; }
    try {
      let value = text;
      if (/^-?\d+$/.test(text)) value = parseInt(text, 10);
      else if (/^-?\d*\.\d+$/.test(text)) value = parseFloat(text);
      await postJSON("/api/update", { table: sql.table, pk, column: col.name, value });
      setCellText(td, value);
      td.classList.remove("saved"); void td.offsetWidth; td.classList.add("saved");
    } catch (e) { setCellText(td, original); toast("Échec : " + e.message); }
  };
  const onKey = (e) => { if (e.key === "Enter") { e.preventDefault(); finish(true); } else if (e.key === "Escape") { e.preventDefault(); finish(false); } };
  const onBlur = () => finish(true);
  td.addEventListener("keydown", onKey);
  td.addEventListener("blur", onBlur);
}

// ── Wiring ─────────────────────────────────────────────────────────────────
$$(".tab").forEach((t) => (t.onclick = () => setView(t.dataset.view)));
$("#filter").addEventListener("input", renderTableList);
$("#prev").onclick = () => { sql.offset = Math.max(0, sql.offset - sql.limit); loadPage(); };
$("#next").onclick = () => { sql.offset += sql.limit; loadPage(); };
$("#save-btn").onclick = async () => {
  const btn = $("#save-btn");
  btn.disabled = true;
  try { await api("/api/save", { method: "POST" }); resetModif(); toast("Sauvegarde écrite · backup .bak créé"); }
  catch (e) { toast("Écriture échouée : " + e.message); }
  finally { btn.disabled = false; }
};

function initLang() {
  const sync = () => $$("#lang button").forEach((b) => b.classList.toggle("active", b.dataset.lang === lang));
  $$("#lang button").forEach((b) => (b.onclick = () => {
    if (b.dataset.lang === lang) return;
    lang = b.dataset.lang;
    localStorage.setItem("dsa-lang", lang);
    sync();
    reloadAll();
  }));
  sync();
}

initLang();
renderHome().catch((e) => toast("Chargement échoué : " + e.message));
