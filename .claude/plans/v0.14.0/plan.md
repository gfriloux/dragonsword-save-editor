## Plan: Cohérence thème de l'UI (point C — polish, T2–T6)

**Type:** UI (web) — polish, aucune logique métier ni format touchés
**Objective:** Éliminer toutes les incohérences visuelles de l'éditeur relevées à
l'audit du point C : plus aucun contrôle « natif navigateur » hors du thème
« Sang & acier ». Concrètement : remplacer les dialogues `prompt()`/`confirm()`
par une modale maison thémée, styler le toggle « verrou », unifier les états
`:focus` et le langage visuel des champs éditables, et thémer les scrollbars.
**Why:** Le point C du backlog vise la cohérence d'édition. L'audit (cette session)
a montré que le rendu propre du thème est cassé par des primitives navigateur non
stylées. Aucune ne touche la sécurité de la save : c'est du polish à faible risque,
mais la **modale maison (T2) est aussi le socle réutilisable** de la future « revue
globale avant écriture » (S3, chantier sécurité séparé).
**Layer(s):** web, docs (aucun `sqlcipher`/`save`/`domain`/`cmd`/`nix`)

---

### Contexte — findings de l'audit (source : `internal/web/static/{app.js,style.css}`)

| Réf | Constat | Emplacement |
|-----|---------|-------------|
| T1  | Spinner natif sur `type="number"` (moche) — **déjà corrigé** | `style.css:233` |
| T2  | Dialogues natifs `prompt()`/`confirm()` (renommage, « Tout débloquer », « Remplir ») — popups OS bruts, hors thème | `app.js:257, 452, 616, 760` |
| T3  | Checkbox « verrou » natif, non stylé | `app.js:642–650`, `.lock` `style.css:257` |
| T4  | États `:focus` incohérents : custom sur stepper/folder-row, absents ailleurs | `style.css:232, 165` vs `252, 309, 355, 413` |
| T5  | Bordures de champs disparates : `--accent` (stepper) vs `--border` (eq-field, inv-fill, sql-rail) | `style.css:229` vs `254, 309, 357` |
| T6  | Scrollbars OS par défaut sur toutes les zones `overflow:auto` (Windows surtout) | global |

Invariants respectés : pur Go / no CGO inchangé (front statique embarqué) ;
aucune écriture de save modifiée ; API inchangée. `DESIGN.md` non impacté sur le
fond (thème = détail d'implémentation de la couche `web`).

---

### Scope

**In:**
- Un mini-composant **modale** en JS vanilla (`app.js`) + son CSS, offrant
  `modalConfirm({title, body, danger})` → `Promise<bool>` et
  `modalPrompt({title, label, value})` → `Promise<string|null>`, thémés,
  fermables au clavier (Échap = annuler, Entrée = valider), focus-trap simple.
- Remplacement des 4 appels natifs `prompt`/`confirm` par ce composant (T2).
- Toggle « verrou » restylé en interrupteur maison (T3), sans changer sa
  sémantique (checkbox sous-jacente conservée pour l'accessibilité).
- Style `:focus` unifié pour **tous** les champs de saisie (T4) et langage
  visuel commun « champ éditable » (T5), via une classe/规则 partagée.
- Scrollbars thémées (T6) : `scrollbar-width`/`scrollbar-color` (Firefox) +
  `::-webkit-scrollbar*` (Chromium/WebKit), aux tokens existants.
- `manual_tests.md` (revue visuelle) + `CHANGELOG.md` (git-cliff).

**Out:**
- Toute logique de sécurité d'écriture (S1 détection jeu ouvert, S2 restauration
  backup, S3 revue globale, S4 bornage) — **chantiers séparés**. Ce plan ne fait
  que **préparer** le socle modale que S3 réutilisera.
- Refonte de la palette / des polices / de la mise en page. On aligne l'existant,
  on ne redessine pas.
- Tout changement d'API, de `domain`, `save`, `sqlcipher`, `cmd`, `nix`.

---

### Files touched
- [ ] `internal/web/static/app.js` (composant modale ; remplacement des 4 appels ; toggle verrou)
- [ ] `internal/web/static/style.css` (modale, toggle, focus, bordures, scrollbars)
- [ ] `internal/web/static/index.html` (conteneur modale si besoin d'un hôte statique)
- [ ] `.claude/plans/v0.14.0/manual_tests.md` (revue visuelle)
- [ ] `README.md` / `DESIGN.md` — **seulement si** une mention UI y devient inexacte (à vérifier ; sinon inchangés)
- [ ] `CHANGELOG.md` (`just changelog`)

---

### Atomic steps

#### Phase 0 — audit
Fait. `nix develop --command just ci` vert (voir `phase0_results.md`).

#### Step 1 — Spinner natif retiré (T1)
Déjà appliqué : règle globale `input[type=number]` dans `style.css` supprimant
les flèches WebKit/Firefox (les boutons ± et raccourcis les remplacent).
**Verification:** `just ci` ; revue visuelle (plus de flèches sur monnaies /
quantité item / remplissage / eq-field).
**Commit:** `fix(web): drop native number-input spinners`

#### Step 2 — Composant modale thémé (T2, socle)
Ajouter dans `app.js` un petit module modale (overlay + carte thémée) exposant
`modalConfirm` et `modalPrompt` renvoyant des `Promise`. CSS dédié aux tokens
« Sang & acier » (surface, bordure, accent), boutons cohérents avec `.cta`
(primaire) et `.cta.ghost` (annuler), variante `danger` (accent rouge) pour les
actions destructrices. Clavier : Échap = annuler, Entrée = valider ; focus initial
sur le champ (prompt) ou le bouton primaire (confirm) ; overlay clic = annuler.
Aucun appel natif remplacé à ce commit (composant introduit isolément).
**Verification:** `just ci` ; test manuel d'ouverture/fermeture (Échap, Entrée, clic overlay).
**Commit:** `feat(web): reusable themed modal (confirm/prompt)`

#### Step 3 — Câbler la modale sur les 4 appels natifs (T2)
Remplacer :
- renommage d'objet (`app.js:257` stepperRow, `:616` namesCell) → `modalPrompt`,
- « Tout débloquer » recettes (`:760`) → `modalConfirm({danger})`,
- « Remplir » catégorie (`:452`) → `modalConfirm`.
Comportements identiques (annulation = pas d'effet). Plus aucun `window.prompt`/
`window.confirm` dans le code.
**Verification:** `just ci` ; test manuel des 4 flux (validation + annulation).
**Commit:** `refactor(web): replace native prompt/confirm with themed modal`

#### Step 4 — Toggle « verrou » thémé (T3)
Restyler `.lock` en interrupteur maison (la `<input type=checkbox>` reste, masquée
visuellement mais focusable ; l'apparence est portée par un pseudo-élément). État
coché = accent. Sémantique et handler `onChange` inchangés.
**Verification:** `just ci` ; test manuel bascule verrou Équipement + Gemmes.
**Commit:** `feat(web): themed lock toggle in equipment/gems`

#### Step 5 — Focus & bordures unifiés (T4 + T5)
Une règle `:focus`/`:focus-visible` commune à tous les champs de saisie (anneau
accent discret, `outline:none` + `box-shadow`/`border-color`), et un langage
visuel « champ éditable » homogène (bordure au repos cohérente : `--border` au
repos, `--accent` au focus, partout). Supprimer les `:focus{outline:none}` isolés
au profit de la règle commune.
**Verification:** `just ci` ; revue visuelle au clavier (Tab) sur home, stepper,
eq-field, inv-fill, sql-rail, cook-search.
**Commit:** `style(web): unified input focus & editable-field borders`

#### Step 6 — Scrollbars thémées (T6)
`scrollbar-width: thin; scrollbar-color: var(--border-2) transparent;` (Firefox)
et `::-webkit-scrollbar`/`-thumb`/`-track` aux tokens sombres, appliqués aux
conteneurs défilants (`.screen`, rails, grilles, détails). Discret, sans capturer
le survol OS natif ailleurs.
**Verification:** `just ci` ; revue visuelle sur les zones défilantes (inventaire,
SQL, cuisine).
**Commit:** `style(web): themed scrollbars`

#### Step 7 — Docs + release
- `manual_tests.md` finalisé (checklist visuelle des 6 items).
- Vérifier `README.md`/`DESIGN.md` : mettre à jour **seulement** si une phrase
  décrivant l'UI devient inexacte (a priori aucune ; le thème est un détail `web`).
- `CHANGELOG.md` via `just changelog`. Gate `just build-windows` (`.exe` statique).
**Commit:** `docs: v0.14.0 theme polish notes` (+ éventuels edits doc en même
commit que le code concerné si une mention bouge).

---

### Technical decisions
- **Modale en JS vanilla**, pas de dépendance : cohérent avec le front actuel
  (zéro build front, tout est dans `app.js`/`style.css` embarqués). Le composant
  est volontairement générique (`confirm`/`prompt`) pour resservir à la revue
  globale S3.
- **Checkbox conservée** sous le toggle stylé (T3) — accessibilité clavier/lecteur
  d'écran gratuite, on ne réinvente pas un widget ARIA.
- **`:focus-visible`** privilégié pour ne pas afficher l'anneau au clic souris.
- **Scrollbars** via propriétés standard + préfixe WebKit ; dégradation propre là
  où non supporté (aucun impact fonctionnel).
- Découpage en commits atomiques par item d'audit : chacun passe `just ci` seul et
  est indépendamment révisable/annulable.

### Risk register
- Risque **faible** (front statique, aucune écriture de save touchée).
- Focus-trap de la modale : garder simple (2–3 éléments focusables) ; ne pas
  bloquer Échap. Testé manuellement.
- Scrollbars WebKit : styliser sans casser le défilement au clavier/molette.
- Régression visuelle possible sur un écran non prévu → la revue visuelle couvre
  tous les écrans (home, inventaire, monnaies, perso, équipe, équipement, gemmes,
  cuisine, SQL).

### Quality gates
- [ ] `nix develop --command just ci` vert à chaque commit
- [ ] `DSA_SAVE=… just test` toujours vert (aucun changement `save`/`domain` attendu)
- [ ] `just build-windows` produit un `.exe` statique unique (CGO_ENABLED=0)
- [ ] Revue visuelle des 9 écrans + les 6 items d'audit (T1–T6) — voir `manual_tests.md`
- [ ] Plus aucun `window.prompt`/`window.confirm` dans `app.js`
- [ ] Commits atomiques sur `feat/theme-polish` ; docs synchronisées si une mention bouge
