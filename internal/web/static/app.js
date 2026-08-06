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

// ── Themed modal (replaces native prompt/confirm) ──────────────────────────
// showModal resolves to a boolean (confirm) or, when `prompt` is true, the
// entered string (Valider) / null (annulé). Keyboard: Enter validates (unless a
// button is focused, which activates itself), Escape and overlay-click cancel;
// Tab is trapped inside the card. Only one modal is expected at a time.
function showModal({ title, message = "", prompt = false, value = "", okText, cancelText = "Annuler" }) {
  return new Promise((resolve) => {
    const cancelVal = prompt ? null : false;
    const overlay = document.createElement("div");
    overlay.className = "modal-overlay";
    const card = document.createElement("div");
    card.className = "modal";
    overlay.appendChild(card);

    if (title) {
      const h = document.createElement("div");
      h.className = "modal-title";
      h.textContent = title;
      card.appendChild(h);
    }
    if (message) {
      const m = document.createElement("div");
      m.className = "modal-msg";
      m.textContent = message;
      card.appendChild(m);
    }
    let field = null;
    if (prompt) {
      field = document.createElement("input");
      field.className = "modal-input";
      field.type = "text";
      field.value = value;
      card.appendChild(field);
    }

    const actions = document.createElement("div");
    actions.className = "modal-actions";
    const cancel = document.createElement("button");
    cancel.className = "modal-btn ghost";
    cancel.textContent = cancelText;
    const ok = document.createElement("button");
    ok.className = "modal-btn primary";
    ok.textContent = okText || (prompt ? "Valider" : "Confirmer");
    actions.append(cancel, ok);
    card.appendChild(actions);

    const focusables = [field, cancel, ok].filter(Boolean);
    const close = (val) => {
      document.removeEventListener("keydown", onKey, true);
      overlay.remove();
      resolve(val);
    };
    const onKey = (e) => {
      if (e.key === "Escape") { e.preventDefault(); close(cancelVal); }
      else if (e.key === "Enter" && e.target.tagName !== "BUTTON") { e.preventDefault(); close(prompt ? field.value.trim() : true); }
      else if (e.key === "Tab") {
        const first = focusables[0], last = focusables[focusables.length - 1];
        if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last.focus(); }
        else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first.focus(); }
      }
    };
    cancel.onclick = () => close(cancelVal);
    ok.onclick = () => close(prompt ? field.value.trim() : true);
    overlay.onclick = (e) => { if (e.target === overlay) close(cancelVal); };
    document.addEventListener("keydown", onKey, true);

    document.body.appendChild(overlay);
    (field || ok).focus();
    if (field) field.select();
  });
}
const modalConfirm = (opts) => showModal({ ...opts, prompt: false });
const modalPrompt = (opts) => showModal({ ...opts, prompt: true });

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

// ── Icons ──────────────────────────────────────────────────────────────────
// Authentic in-game icons come from /api/icon?cid=…; on failure (no game folder,
// no icon, extraction error) they fall back to the th.gl sprite atlas, then a
// category dot.
let iconSize = 64;
const ICON_DISP = 40;
const spriteStyle = (it, disp) =>
  `object-fit:none;object-position:${it.iconX}px ${it.iconY}px;` +
  `width:${iconSize}px;height:${iconSize}px;zoom:${disp / iconSize}`;

// iconFail swaps a failed authentic <img> to its sprite cell, or a dot.
function iconFail(img) {
  img.onerror = null;
  if (img.dataset.sprite === "1") {
    img.src = "/sprite.webp";
    img.style.cssText =
      `object-fit:none;object-position:${img.dataset.x}px ${img.dataset.y}px;` +
      `width:${iconSize}px;height:${iconSize}px;zoom:${+img.dataset.disp / iconSize}`;
  } else {
    const s = document.createElement("span");
    s.className = "cat-dot";
    img.replaceWith(s);
  }
}
window.iconFail = iconFail;

function iconEl(it, disp = ICON_DISP) {
  if (!it || !it.cid) {
    const d = document.createElement("span");
    d.className = "cat-dot";
    return d;
  }
  const img = document.createElement("img");
  img.className = "ic";
  img.alt = "";
  img.width = img.height = disp;
  img.style.objectFit = "contain";
  img.src = "/api/icon?cid=" + it.cid;
  img.dataset.sprite = it.icon ? "1" : "0";
  img.dataset.x = it.iconX || 0;
  img.dataset.y = it.iconY || 0;
  img.dataset.disp = disp;
  img.onerror = () => iconFail(img);
  return img;
}
const iconHTML = (it, disp = ICON_DISP) =>
  it && it.cid
    ? `<img class="ic" alt="" width="${disp}" height="${disp}" style="object-fit:contain" ` +
      `src="/api/icon?cid=${it.cid}" data-sprite="${it.icon ? 1 : 0}" ` +
      `data-x="${it.iconX || 0}" data-y="${it.iconY || 0}" data-disp="${disp}" onerror="iconFail(this)">`
    : `<span class="cat-dot"></span>`;

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
    edit.onclick = async () => {
      const v = await modalPrompt({ title: "Renommer", message: `Nom pour CID ${cid}`, value: known ? name : "" });
      if (v !== null) onLabel(v);
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
let selInvKey = null; // selected cell (detail panel)
let invBaseline = {}; // key -> value read from the save (for the amber "modified" dot)
const catLabel = (c) => (lang === "fr" ? c.labelFr : c.labelEn) || c.labelEn || c.key;

const RARITY = [
  { g: "normal", fr: "Commun", en: "Common" },
  { g: "rare", fr: "Rare", en: "Rare" },
  { g: "superior", fr: "Supérieur", en: "Superior" },
  { g: "epic", fr: "Épique", en: "Epic" },
  { g: "legendary", fr: "Légendaire", en: "Legendary" },
];
const entryKey = (e) => (e.stackable ? "c" + e.cid : "k" + (e.item.id || e.cid));
const rarityClass = (e) => (e.item && e.item.grade ? "r-" + e.item.grade : "r-normal");

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
  // Capture the save baseline once per item (drives the amber "modified" dot).
  for (const k in model.entries) for (const e of model.entries[k]) {
    const key = entryKey(e);
    if (invBaseline[key] === undefined) invBaseline[key] = e.value;
    e.saved = invBaseline[key];
  }
  const cats = model.cats.filter((c) => (model.entries[c.key] || []).length);

  const rail = document.createElement("aside");
  rail.className = "inv-rail";
  const main = document.createElement("div");
  main.className = "inv-main";
  const detail = document.createElement("div");
  detail.className = "inv-detail empty";
  el.append(rail, main, detail);

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
      selInvKey = null;
      $$(".cat-link").forEach((x) => x.classList.toggle("active", x.dataset.key === c.key));
      renderInvCat(main, detail, model, c);
    };
    rail.appendChild(b);
  }
  renderInvCat(main, detail, model, cats.find((c) => c.key === selectedCat));
};

function renderInvCat(main, detail, model, cat) {
  const list = model.entries[cat.key] || [];
  main.innerHTML = "";
  const bar = document.createElement("div");
  bar.className = "inv-toolbar";
  const ownedN = () => list.filter((e) => e.owned).length;
  const count = document.createElement("span");
  count.className = "count";
  const refreshCount = () => (count.textContent = `${ownedN()} possédés / ${list.length}`);
  const h = document.createElement("h2");
  h.textContent = catLabel(cat);
  bar.append(h, count);
  refreshCount();
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
    if (!(await modalConfirm({ title: "Remplir la catégorie", message: `Mettre les ${stacks.length} objets empilables de « ${catLabel(cat)} » à ${n} ?`, okText: "Remplir" }))) return;
    try { for (const e of stacks) await postJSON("/api/game/stackable", { cid: e.cid, count: n }); await RENDER.inv(); }
    catch (err) { toast("Remplissage échoué : " + err.message); }
  };
  fillWrap.append(fillN, fillBtn);
  bar.appendChild(fillWrap);
  main.appendChild(bar);

  const grid = document.createElement("div");
  grid.className = "inv-grid";
  const cellOf = {};
  for (const e of list) {
    const key = entryKey(e);
    const cell = document.createElement("div");
    cell.className = "cell " + rarityClass(e);
    cell.appendChild(iconEl(e.item, 56));
    const q = document.createElement("span");
    q.className = "badge-q";
    const dot = document.createElement("span");
    dot.className = "dirty";
    cell.append(q, dot);
    const paint = () => {
      q.textContent = e.value;
      cell.classList.toggle("empty", e.value <= 0);
      dot.style.display = e.value !== e.saved ? "block" : "none";
    };
    paint();
    cell._paint = paint;
    cell.onclick = () => selectInvCell(detail, model, cat, e, cell, grid, refreshCount);
    cellOf[key] = cell;
    grid.appendChild(cell);
  }
  main.appendChild(grid);

  // Keep or reset the detail selection.
  const keep = list.find((e) => entryKey(e) === selInvKey);
  if (keep) selectInvCell(detail, model, cat, keep, cellOf[selInvKey], grid, refreshCount);
  else { detail.className = "inv-detail empty"; detail.innerHTML = "<span>Choisis un objet.</span>"; }
}

function selectInvCell(detail, model, cat, e, cell, grid, refreshCount) {
  selInvKey = entryKey(e);
  grid.querySelectorAll(".cell").forEach((c) => c.classList.remove("sel"));
  if (cell) cell.classList.add("sel");
  renderInvDetail(detail, e, cell, refreshCount);
}

function renderInvDetail(detail, e, cell, refreshCount) {
  detail.className = "inv-detail";
  detail.innerHTML = "";
  const ic = document.createElement("div");
  ic.className = "detail-ic";
  ic.appendChild(iconEl(e.item, 96));
  const name = document.createElement("div");
  name.className = "detail-name" + (e.known ? "" : " unknown");
  name.textContent = e.name;
  const meta = document.createElement("div");
  meta.className = "icid";
  meta.textContent = `CID ${e.cid} · ${e.stackable ? "tb_stackable_item" : "tb_cook"}`;
  const sep = document.createElement("div");
  sep.className = "detail-sep";
  const lbl = document.createElement("div");
  lbl.className = "label";
  lbl.textContent = "Quantité en stock";
  detail.append(ic, name, meta, sep, lbl);

  const commit = async (v) => {
    v = Math.max(0, Math.min(9999, v | 0));
    if (v === e.value) return;
    try {
      await e.commit(v);
      e.value = v; e.owned = v > 0;
      if (cell && cell._paint) cell._paint();
      refreshCount();
      renderInvDetail(detail, e, cell, refreshCount);
    } catch (err) { toast("Échec : " + err.message); }
  };
  detail.appendChild(stepper(e.value, 1, commit));

  const presets = document.createElement("div");
  presets.className = "presets";
  for (const p of [0, 99, 999, 9999]) {
    const b = document.createElement("button");
    if (p === 9999) b.className = "max";
    b.textContent = p === 9999 ? "MAX" : p;
    b.onclick = () => commit(p);
    presets.appendChild(b);
  }
  detail.appendChild(presets);

  if (e.value !== e.saved) {
    const diff = document.createElement("div");
    diff.className = "diff-inset";
    const undo = document.createElement("a");
    undo.textContent = "Annuler cette modification";
    undo.onclick = () => commit(e.saved);
    diff.innerHTML = `<span class="t">Modifié · non écrit</span><span class="v">${e.saved} → ${e.value}</span>`;
    diff.appendChild(undo);
    detail.appendChild(diff);
  }

  const legend = document.createElement("div");
  legend.className = "legend";
  for (const r of RARITY) {
    legend.innerHTML += `<span class="lg"><span class="sw" style="border-color:var(--r-${r.g === "normal" ? "common" : r.g})"></span>${lang === "fr" ? r.fr : r.en}</span>`;
  }
  detail.appendChild(legend);
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
    edit.onclick = async () => { const v = await modalPrompt({ title: "Renommer", message: `Nom pour CID ${cid}`, value: known ? name : "" }); if (v !== null) onLabel(v); };
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
  const sw = document.createElement("span");
  sw.className = "switch";
  cb.onchange = async () => { try { await onChange(cb.checked); } catch (e) { cb.checked = !cb.checked; toast("Échec : " + e.message); } };
  label.append(cb, sw, document.createTextNode("verrou"));
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

// ── Costumes & Familiers (cosmetics + mounts) ──────────────────────────────
// A themed <select> that reflects a current value and posts on change, rolling
// back to the previous choice on failure.
function equipSelect(options, current, onCommit) {
  const sel = document.createElement("select");
  sel.className = "equip-select";
  for (const [val, label] of options) {
    const o = document.createElement("option");
    o.value = val; o.textContent = label;
    if (String(val) === String(current)) o.selected = true;
    sel.appendChild(o);
  }
  let prev = sel.value;
  sel.onchange = async () => {
    try { await onCommit(sel.value); prev = sel.value; sel.classList.remove("saved"); void sel.offsetWidth; sel.classList.add("saved"); }
    catch (e) { sel.value = prev; toast("Échec : " + e.message); }
  };
  return sel;
}

function sectionTitle(text) {
  const t = document.createElement("div");
  t.className = "group-title";
  t.textContent = text;
  return t;
}

// unlockGrid renders not-owned catalog entries as cards, each with an unlock
// button; returns the grid, or an empty-note paragraph when there is nothing.
function unlockGrid(entries, onUnlock, refresh, emptyMsg) {
  if (!entries.length) {
    const p = document.createElement("p");
    p.className = "empty-note";
    p.textContent = emptyMsg;
    return p;
  }
  const grid = document.createElement("div");
  grid.className = "cards";
  for (const e of entries) {
    const card = document.createElement("div");
    card.className = "card";
    const head = document.createElement("div");
    head.className = "card-head";
    head.appendChild(iconEl(e, 40));
    head.appendChild(namesCell(displayName(e), e.cid, e.known, null));
    card.appendChild(head);
    const btn = document.createElement("button");
    btn.className = "action-btn";
    btn.textContent = "Débloquer";
    btn.onclick = async () => {
      btn.disabled = true;
      try { await onUnlock(e.cid); toast("Débloqué — clique « Écrire dans la save »."); await refresh(); }
      catch (err) { btn.disabled = false; toast("Échec : " + err.message); }
    };
    card.appendChild(btn);
    grid.appendChild(card);
  }
  return grid;
}

RENDER.costumes = async function () {
  const el = $("#screen-costumes");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Garde-robe</span><h2>Costumes</h2><span class="sub">Tenues et skins d'arme (<span class="mono">tb_costume</span>). Un personnage peut en porter plusieurs à la fois. Débloque et équipe librement.</span></div>`;
  const { costumes, catalog, characters } = await api("/api/game/costumes");

  const charOpts = [[0, "— non porté —"], ...(characters || []).map((c) => [c.cid, displayName(c)])];
  const owned = document.createElement("div");
  owned.className = "rows";
  if (!(costumes && costumes.length)) owned.innerHTML = `<p class="empty-note">Aucun costume possédé.</p>`;
  for (const c of costumes || []) {
    const row = document.createElement("div");
    row.className = "row";
    row.appendChild(iconEl(c, 32));
    row.appendChild(namesCell(displayName(c), c.cid, c.known, (name) => postJSON("/api/game/label", { cid: c.cid, name }).then(() => RENDER.costumes())));
    const fields = document.createElement("div");
    fields.className = "eq-fields";
    const lbl = document.createElement("label");
    lbl.className = "eq-field";
    lbl.innerHTML = "<span>Porté par</span>";
    lbl.appendChild(equipSelect(charOpts, c.equipCharacterCid, (v) => postJSON("/api/game/costumes/equip", { dbid: c.dbid, characterCid: Number(v) })));
    fields.appendChild(lbl);
    row.appendChild(fields);
    owned.appendChild(row);
  }

  const locked = (catalog || []).filter((e) => !e.owned);
  el.appendChild(sectionTitle("Possédés"));
  el.appendChild(owned);
  el.appendChild(sectionTitle(`Débloquer (${locked.length})`));
  el.appendChild(unlockGrid(locked, (cid) => postJSON("/api/game/costumes/unlock", { cid }), () => RENDER.costumes(), "Tous les costumes sont déjà possédés."));
};

RENDER.familiers = async function () {
  const el = $("#screen-familiers");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Écurie</span><h2>Familiers</h2><span class="sub">Créatures montables (<span class="mono">tb_vehicle</span>). Une monture par personnage. Débloque et assigne à chaque héros.</span></div>`;
  const { vehicles, catalog, mounts, characters } = await api("/api/game/familiers");

  const mountByChar = {};
  for (const m of mounts || []) mountByChar[m.characterCid] = m;
  const vehOpts = [[0, "— aucune —"], ...(vehicles || []).map((v) => [v.dbid, displayName(v)])];

  const rows = document.createElement("div");
  rows.className = "rows";
  if (!(characters && characters.length)) rows.innerHTML = `<p class="empty-note">Aucun personnage.</p>`;
  for (const ch of characters || []) {
    const cur = mountByChar[ch.cid];
    const row = document.createElement("div");
    row.className = "row";
    row.appendChild(iconEl(ch, 32));
    row.appendChild(namesCell(displayName(ch), ch.cid, ch.known, null));
    const fields = document.createElement("div");
    fields.className = "eq-fields";
    const lbl = document.createElement("label");
    lbl.className = "eq-field";
    lbl.innerHTML = "<span>Monture</span>";
    lbl.appendChild(equipSelect(vehOpts, cur ? cur.vehicleDbid : 0, (v) => postJSON("/api/game/familiers/equip", { characterCid: ch.cid, vehicleDbid: v })));
    fields.appendChild(lbl);
    row.appendChild(fields);
    rows.appendChild(row);
  }

  const locked = (catalog || []).filter((e) => !e.owned);
  el.appendChild(sectionTitle("Montures par personnage"));
  el.appendChild(rows);
  el.appendChild(sectionTitle(`Débloquer (${locked.length})`));
  el.appendChild(unlockGrid(locked, (cid) => postJSON("/api/game/familiers/unlock", { cid }), () => RENDER.familiers(), "Tous les familiers sont déjà possédés."));
};

// ── Titres (flat checklist, bitmask tb_title) ──────────────────────────────
const titles = { list: [], search: "", filter: "all" };

const TITLE_COLORS = { BLUE: "#5aa0e0", BROWN: "#b08050", GREEN: "#6fcf7f", RED: "#e05a5a", VIOLET: "#b98ce0" };
const STAT_LABELS = {
  MaxHP: ["PV max", "Max HP"], Attack: ["Attaque", "Attack"], Defence: ["Défense", "Defence"],
};

function titleName(t) {
  const n = lang === "fr" ? t.nameFr : t.nameEn;
  return n || t.nameEn || t.nameFr || "ID " + t.id;
}

function statText(t) {
  return (t.stats || [])
    .map((s) => `${(STAT_LABELS[s.type] || [s.type, s.type])[lang === "fr" ? 0 : 1]} +${s.value}`)
    .join(" · ");
}

RENDER.titles = async function () {
  const el = $("#screen-titles");
  el.className = "screen panel";
  el.innerHTML = `<div class="panel-head"><span class="overline">Distinctions</span><h2>Titres</h2><span class="sub">Titres de compte (<span class="mono">tb_title</span>, masque de bits). Coche pour débloquer ; le titre affiché en jeu n'est pas modifié.</span></div>`;
  const data = await api("/api/game/titles");
  titles.list = data.titles || [];

  const bar = document.createElement("div");
  bar.className = "titles-toolbar";
  const count = document.createElement("span");
  count.className = "cook-count";
  titles._count = count;

  const filterSeg = document.createElement("div");
  filterSeg.className = "seg";
  for (const [val, label] of [["all", "Tous"], ["locked", "À débloquer"], ["unlocked", "Débloqués"]]) {
    const b = document.createElement("button");
    b.textContent = label;
    b.className = titles.filter === val ? "on" : "";
    b.onclick = () => {
      titles.filter = val;
      filterSeg.querySelectorAll("button").forEach((x) => x.classList.toggle("on", x === b));
      paintTitles();
    };
    filterSeg.appendChild(b);
  }

  const search = document.createElement("input");
  search.className = "cook-search";
  search.type = "search";
  search.placeholder = "Rechercher un titre…";
  search.value = titles.search;
  search.oninput = () => { titles.search = search.value; paintTitles(); };

  const unlock = document.createElement("button");
  unlock.className = "action-btn";
  unlock.textContent = "Tout débloquer";
  unlock.onclick = async () => {
    if (!(await modalConfirm({ title: "Tout débloquer", message: "Débloquer les 108 titres ?", okText: "Tout débloquer" }))) return;
    unlock.disabled = true;
    try {
      await postJSON("/api/game/titles/unlock-all", {});
      for (const t of titles.list) t.unlocked = true;
      paintTitles();
      toast("Titres débloqués — clique « Écrire dans la save ».");
    } catch (e) { toast("Échec : " + e.message); }
    finally { unlock.disabled = false; }
  };

  bar.append(count, filterSeg, search, unlock);
  el.appendChild(bar);

  const rows = document.createElement("div");
  rows.className = "rows";
  titles._rows = rows;
  el.appendChild(rows);
  paintTitles();
};

function paintTitles() {
  const q = titles.search.trim().toLowerCase();
  const shown = titles.list.filter((t) => {
    if (titles.filter === "locked" && t.unlocked) return false;
    if (titles.filter === "unlocked" && !t.unlocked) return false;
    if (q && !titleName(t).toLowerCase().includes(q)) return false;
    return true;
  });
  const on = titles.list.filter((t) => t.unlocked).length;
  if (titles._count) titles._count.textContent = `${on}/${titles.list.length} débloqués`;

  const rows = titles._rows;
  rows.innerHTML = "";
  if (!shown.length) { rows.innerHTML = `<p class="empty-note">Aucun titre.</p>`; return; }
  for (const t of shown) {
    const row = document.createElement("label");
    row.className = "row title-row";

    const cb = document.createElement("input");
    cb.type = "checkbox";
    cb.checked = t.unlocked;
    cb.className = "title-check";
    cb.onchange = async () => {
      cb.disabled = true;
      try {
        await postJSON("/api/game/titles/unlock", { id: t.id, unlocked: cb.checked });
        t.unlocked = cb.checked;
        if (titles._count) titles._count.textContent = `${titles.list.filter((x) => x.unlocked).length}/${titles.list.length} débloqués`;
        if (titles.filter !== "all") paintTitles();
      } catch (e) { cb.checked = !cb.checked; toast("Échec : " + e.message); }
      finally { cb.disabled = false; }
    };

    const dot = document.createElement("span");
    dot.className = "title-dot";
    dot.style.background = TITLE_COLORS[t.color] || "var(--muted)";
    dot.title = t.color || "";

    const names = document.createElement("div");
    names.className = "names";
    const iname = document.createElement("span");
    iname.className = "iname" + (t.unlocked ? "" : " unknown");
    iname.textContent = titleName(t);
    const icid = document.createElement("span");
    icid.className = "icid";
    icid.textContent = "ID " + t.id;
    names.append(iname, icid);

    row.append(cb, dot, names);
    const st = statText(t);
    if (st) {
      const stats = document.createElement("span");
      stats.className = "title-stats";
      stats.textContent = st;
      row.appendChild(stats);
    }
    rows.appendChild(row);
  }
}

// ── Cuisine (recipe grid + detail panel) ───────────────────────────────────
const cook = { recipes: [], tools: [], tool: "", search: "", known: "all", sel: null };

function cookToolLabel(type) {
  const t = cook.tools.find((x) => x.type === type);
  return t ? (lang === "fr" ? t.labelFr : t.labelEn) : type;
}
function localized(o) { return o ? (lang === "fr" ? o.fr : o.en) : ""; }
function dishName(dish, key) {
  return (lang === "fr" ? dish.nameFr : dish.nameEn) || `Plat ${dish.cid || key}`;
}

RENDER.cook = async function () {
  const el = $("#screen-cook");
  el.className = "screen cook";
  el.innerHTML = "";

  const data = await api("/api/game/recipes");
  cook.recipes = data.recipes || [];
  cook.tools = data.tools || [];
  if (!cook.recipes.some((r) => r.key === cook.sel)) cook.sel = null;

  const main = document.createElement("div");
  main.className = "cook-main";
  const detail = document.createElement("aside");
  detail.className = "cook-detail empty";
  detail.innerHTML = "<span>Choisis une recette.</span>";
  el.append(main, detail);
  cook._detail = detail;

  // Toolbar: title + count, then filters.
  const bar = document.createElement("div");
  bar.className = "cook-toolbar";
  const head = document.createElement("div");
  head.className = "cook-head";
  head.innerHTML = `<div><span class="overline">Fourneaux</span><h2>Livre de recettes</h2></div>`;
  const count = document.createElement("span");
  count.className = "cook-count";
  head.appendChild(count);
  cook._count = count;

  const controls = document.createElement("div");
  controls.className = "cook-controls";
  const toolSeg = segGroup("tool", [["", "Tous"], ...cook.tools.map((t) => [t.type, lang === "fr" ? t.labelFr : t.labelEn])]);
  const knownSeg = segGroup("known", [["all", "Toutes"], ["unknown", "Inconnues"], ["known", "Connues"]]);
  const search = document.createElement("input");
  search.type = "search";
  search.className = "cook-search";
  search.placeholder = "Rechercher un plat…";
  search.value = cook.search;
  search.oninput = () => { cook.search = search.value; paintCook(); };
  const unlock = document.createElement("button");
  unlock.className = "action-btn";
  unlock.textContent = "Tout débloquer";
  unlock.onclick = async () => {
    if (!(await modalConfirm({ title: "Tout débloquer", message: "Marquer toutes les recettes comme connues ?", okText: "Tout débloquer" }))) return;
    unlock.disabled = true;
    try {
      await postJSON("/api/game/recipes/unlock-all", {});
      for (const r of cook.recipes) r.known = true;
      paintCook();
      const sel = cook.recipes.find((r) => r.key === cook.sel);
      if (sel) renderCookDetail(sel);
      toast("Recettes débloquées — clique « Écrire dans la save ».");
    } catch (e) { toast("Échec : " + e.message); }
    finally { unlock.disabled = false; }
  };
  controls.append(toolSeg, knownSeg, search, unlock);
  bar.append(head, controls);

  const grid = document.createElement("div");
  grid.className = "cook-grid";
  main.append(bar, grid);
  cook._grid = grid;

  paintCook();
  const sel = cook.recipes.find((r) => r.key === cook.sel);
  if (sel) renderCookDetail(sel);
};

function segGroup(key, options) {
  const seg = document.createElement("div");
  seg.className = "seg";
  for (const [val, label] of options) {
    const b = document.createElement("button");
    b.textContent = label;
    b.className = cook[key] === val ? "on" : "";
    b.onclick = () => {
      cook[key] = val;
      seg.querySelectorAll("button").forEach((x) => x.classList.toggle("on", x === b));
      paintCook();
    };
    seg.appendChild(b);
  }
  return seg;
}

function cookVisible() {
  const q = cook.search.trim().toLowerCase();
  return cook.recipes.filter((r) => {
    if (cook.tool && r.tool !== cook.tool) return false;
    if (cook.known === "known" && !r.known) return false;
    if (cook.known === "unknown" && r.known) return false;
    if (q) {
      const n = ((r.dish && (r.dish.nameFr + " " + r.dish.nameEn)) || "").toLowerCase();
      if (!n.includes(q)) return false;
    }
    return true;
  });
}

function paintCook() {
  const vis = cookVisible();
  const knownN = cook.recipes.filter((r) => r.known).length;
  cook._count.textContent = `${knownN} / ${cook.recipes.length} connues`;

  const grid = cook._grid;
  grid.innerHTML = "";
  if (!vis.length) {
    grid.innerHTML = `<p class="empty-note">Aucune recette ne correspond.</p>`;
    return;
  }
  const frag = document.createDocumentFragment();
  for (const r of vis) frag.appendChild(cookCard(r));
  grid.appendChild(frag);
}

function cookCard(r) {
  const dish = r.dish || {};
  const card = document.createElement("button");
  card.className = "cook-card " + (r.known ? "known" : "locked") + (cook.sel === r.key ? " sel" : "");
  card.dataset.key = r.key;
  const vis = document.createElement("div");
  vis.className = "cook-card-vis";
  vis.appendChild(iconEl(dish, 64));
  const badge = document.createElement("span");
  badge.className = "cook-badge " + (r.known ? "known" : "locked");
  badge.textContent = r.known ? "✓" : "";
  vis.appendChild(badge);
  const nm = document.createElement("span");
  nm.className = "cook-card-name";
  nm.textContent = dishName(dish, r.key);
  card.append(vis, nm);
  card.onclick = () => selectCook(r);
  return card;
}

function selectCook(r) {
  cook.sel = r.key;
  cook._grid.querySelectorAll(".cook-card").forEach((c) => c.classList.toggle("sel", Number(c.dataset.key) === r.key));
  renderCookDetail(r);
}

function renderCookDetail(r) {
  const d = cook._detail;
  d.className = "cook-detail";
  d.innerHTML = "";
  const dish = r.dish || {};

  const vis = document.createElement("div");
  vis.className = "cook-detail-vis";
  vis.appendChild(iconEl(dish, 120));
  const name = document.createElement("div");
  name.className = "cook-detail-name" + (r.known ? "" : " unknown");
  name.textContent = dishName(dish, r.key);
  const tool = document.createElement("div");
  tool.className = "cook-detail-tool";
  tool.textContent = cookToolLabel(r.tool);
  d.append(vis, name, tool);

  if (r.effects && r.effects.length) {
    const eff = document.createElement("div");
    eff.className = "cook-effect";
    const ov = document.createElement("div");
    ov.className = "overline";
    ov.textContent = "Effet" + (r.effectName ? " · " + localized(r.effectName) : "");
    const txt = document.createElement("div");
    txt.className = "cook-effect-txt";
    txt.textContent = localized(r.effects[0]);
    eff.append(ov, txt);
    const distinct = [...new Set(r.effects.map(localized).filter(Boolean))];
    if (distinct.length > 1) {
      const scale = document.createElement("div");
      scale.className = "cook-effect-scale";
      scale.textContent = "Augmente avec la qualité (5 paliers)";
      scale.title = distinct.join("\n");
      eff.appendChild(scale);
    }
    d.appendChild(eff);
  }

  const mHead = document.createElement("div");
  mHead.className = "overline mat-head";
  mHead.textContent = "Matériaux requis";
  d.appendChild(mHead);
  const mats = document.createElement("div");
  mats.className = "cook-mats";
  for (const g of r.ingredients || []) mats.appendChild(matTile(g));
  d.appendChild(mats);

  const seg = document.createElement("div");
  seg.className = "seg cook-known-seg";
  const mk = (known, label) => {
    const b = document.createElement("button");
    b.textContent = label;
    b.className = r.known === known ? "on" : "";
    b.onclick = async () => {
      if (r.known === known) return;
      seg.querySelectorAll("button").forEach((x) => (x.disabled = true));
      try {
        await postJSON("/api/game/recipes/known", { key: r.key, known });
        r.known = known;
        paintCook();
        renderCookDetail(r);
        toast(known ? "Recette connue — écris la save." : "Recette verrouillée — écris la save.");
      } catch (e) { toast("Échec : " + e.message); seg.querySelectorAll("button").forEach((x) => (x.disabled = false)); }
    };
    return b;
  };
  seg.append(mk(false, "Verrouillée"), mk(true, "Connue"));
  d.appendChild(seg);

  const ref = document.createElement("div");
  ref.className = "cook-ref";
  ref.textContent = `tb_switch · cat ${r.category} · bit ${r.bit}`;
  d.appendChild(ref);
}

function matTile(g) {
  const tile = document.createElement("div");
  const have = g.owned >= g.qty;
  tile.className = "mat-tile " + (have ? "have" : "miss");
  const vis = document.createElement("div");
  vis.className = "mat-vis";
  if (g.iconCid) {
    vis.appendChild(iconEl({ cid: g.iconCid, iconPath: g.iconPath, icon: false }, 36));
  } else {
    const dot = document.createElement("span");
    dot.className = "mat-dot";
    vis.appendChild(dot);
  }
  const lbl = document.createElement("div");
  lbl.className = "mat-lbl";
  const name = (lang === "fr" ? g.nameFr : g.nameEn) || String(g.id);
  lbl.textContent = g.qty > 1 ? `${name} ×${g.qty}` : name;
  const ratio = document.createElement("div");
  ratio.className = "mat-ratio";
  ratio.textContent = `${g.owned}/${g.qty}`;
  tile.append(vis, lbl, ratio);
  tile.title = g.kind === "type" ? "Catégorie d'ingrédient (total possédé dans la save)" : "Ingrédient précis";
  return tile;
}

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
  try {
    await api("/api/save", { method: "POST" });
    resetModif();
    invBaseline = {}; // written to disk → new baseline
    // Refresh the inventory baseline (amber dots), but never un-hide it:
    // RENDER.inv() rebuilds el.className and would strip the "hidden" class.
    if (currentView === "inv") RENDER.inv().catch(() => {});
    else loaded.delete("inv"); // recapture baseline on next view
    toast("Sauvegarde écrite · backup .bak créé");
  } catch (e) { toast("Écriture échouée : " + e.message); }
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
