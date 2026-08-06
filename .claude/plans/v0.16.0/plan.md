## Plan: Titres — un nouvel écran joueur (déblocage data-driven)

**Type:** edit capability + UI + cmd (converter offline) — sur le patron Cuisine (recettes)
**Objective:** Un panneau « Titres » qui liste **les 108 titres du jeu** résolus en nom Fr/En,
affiche lesquels sont débloqués, et permet de **débloquer** un titre (case à cocher) ou **tout
débloquer** d'un clic. Table `tb_title` aujourd'hui accessible seulement en SQL brut.
**Why:** BACKLOG (`Titres — tb_title. Déblocage.`), priorité haute, risque faible (compte-joueur,
pas de progression corrompue). Modèle bitmask **identique aux recettes** (livré v0.13), catalogue
**data-driven** extrait des paks comme les items/recettes (v0.14). Étude complète le 2026-08-06
(voir « Modèle confirmé » ci-dessous et la mémoire `dsa-titles`).
**Layer(s):** cmd (converter offline), domain, web, docs. Aucun changement `sqlcipher` / `save` /
`pak` / `oodle`.

---

### Contexte — modèle confirmé (vrai save `6144_Slot1.db` + extraction pak, 2026-08-06)

**Table `tb_title`** `PRIMARY KEY (USER_DBID, CATEGORY)` :
- `BIT_FIELD` INTEGER — masque 64-bit des titres **débloqués** de la catégorie.
- `FAV_BIT_FIELD` INTEGER — masque des titres favoris / affiché(s). **Hors scope** (voir Out).

**Invariant** (vérifié : `32812*64+36 = 2100004`, `32875*64+7 = 2104007`) :
```
title_id = CATEGORY * 64 + bit   →   CATEGORY = id / 64 ,  bit = id % 64
```
Débloquer un titre = allumer son bit dans `BIT_FIELD`, en créant la ligne de catégorie si absente
(`INSERT OR REPLACE`, read-modify-write). Mécaniquement **identique à `SetRecipeKnown` /
`UnlockAllRecipes`** (`internal/domain/recipes.go`).

**Catalogue maître = `AccountTitleData.xml`** (extrait des paks, sauvegardé sous
`tmp/pak/AccountTitleData.xml`, gitignored — chemin pak :
`DS/Content/__GeneratedGameData__/Server/XML/GameData/AccountTitleData.xml`) :
- **108 titres**, IDs `2100000`–`2104xxx`, **7 catégories** : 32812, 32828, 32843, 32844, 32859,
  32875, 32876.
- Par ligne : `ID`, `Name` (clé StringData, ex. `902100004`), `Memo` (nom dev coréen),
  `FontColor` (BLUE/BROWN/GREEN/RED/VIOLET), `AchievementGroupID`/`AchievementStep` (comment
  l'obtenir), `StatType1..3`/`StatValue1..3` (bonus de stats accordé : MaxHP/Attack/Defence).
- Noms résolus via `tmp/pak/StringData.xml` (Fr + En natifs), ex. `2100004` → « Mains curieuses »
  / « Curious Hands ». Les 6 titres débloqués du fixture existent tous dans le master ✓.

**Point d'implémentation** : la catégorie 32843 a des titres sur les bits hauts (masque > 2⁶³) →
manipuler `BIT_FIELD` en `uint64` puis stocker `int64(field)` (complément à deux), exactement comme
`recipes.go` le fait déjà (`int64(uint64(cur) | mask)`). On ne pose **que** les 108 bits réels,
jamais un masque plein `0xFFFF…`.

---

### Scope

**In:**
- **cmd (offline)** : `cmd/pak-titles` (pur Go) lit `AccountTitleData.xml` + `StringData.xml`
  déjà extraits (fixtures `tmp/pak/`) → émet `internal/domain/data/titles.json` (id, nom Fr/En,
  couleur, bonus de stats). Même patron offline que `cmd/pak-catalog`.
- **domain (lecture)** : type `Title` ; `Titles()` = les 108 titres résolus avec l'état *débloqué*
  lu depuis `tb_title.BIT_FIELD`.
- **domain (écriture)** : `SetTitleUnlocked(id, unlocked)` (read-modify-write d'un bit, no-op idempotent)
  et `UnlockAllTitles()` (OR des bits réels par catégorie, transactionnel).
- **web** : un panneau « Titres » + endpoints `/api/game/titles`, `/api/game/titles/unlock`,
  `/api/game/titles/unlock-all`.
- **docs** : `docs/database.md` (schéma + sémantique `tb_title`), `docs/content-ids.md` (plage
  `2100000`–`2104xxx`, 7 catégories), `docs/switches.md` (note : `tb_title` est de la même famille
  bitmask), `DESIGN.md` / `README.md` (nouvel écran), `CHANGELOG` via git-cliff. Doc du converter
  dans l'en-tête de `cmd/pak-titles` + `tmp/pak/README.md` (fichier titres ajouté à la recette).

**Out:**
- **`FAV_BIT_FIELD`** (titre équipé / affiché) : sémantique non vérifiée (mono-sélection ? un par
  catégorie ?). Déblocage uniquement pour cette version ; documenté comme extension future.
- **Re-verrouiller en masse** : `SetTitleUnlocked(id,false)` existe pour la symétrie de la case,
  mais aucun « tout retirer » (faible valeur, risque de confusion).
- Édition des **bonus de stats** ou du `FontColor` (données figées du jeu).
- Résolution des **icônes de titre** (aucune correspondance id→sprite confirmée ; on affiche du
  texte + une pastille de couleur).
- Tout changement `sqlcipher` / `save` / `pak` / `oodle`, et l'extraction pak elle-même (offline,
  déjà faite ; l'artefact vit sous `tmp/`).

---

### Files touched
- [ ] `cmd/pak-titles/main.go` — converter offline XML → `data/titles.json`
- [ ] `internal/domain/data/titles.json` — catalogue embarqué (108 titres)
- [ ] `internal/domain/catalog.go` — ajouter `data/titles.json` à la directive `//go:embed`
- [ ] `internal/domain/title.go` — type `Title` + `Titles()`, `SetTitleUnlocked`, `UnlockAllTitles`
- [ ] `internal/domain/title_test.go` — lecture + round-trip unlock sur vrai save (`DSA_SAVE`)
- [ ] `internal/web/game.go` — handlers `handleTitles`, `handleTitleUnlock`, `handleUnlockTitles`
- [ ] `internal/web/web.go` — enregistrement des routes
- [ ] `internal/web/static/{index.html,app.js,style.css}` — panneau Titres
- [ ] `docs/database.md`, `docs/content-ids.md`, `docs/switches.md`, `DESIGN.md`, `README.md`
- [ ] `.claude/plans/v0.16.0/manual_tests.md`

---

### Atomic steps

#### Step 0: Audit Phase 0
**Description:** `nix develop --command just ci` sur `main` propre (dernier commit `6ad606b`),
résultat consigné dans `phase0_results.md`. Créer la branche `feat/titles`.
**Verification:** `just ci` vert avant toute ligne de code.
**Commit:** — (pas de commit ; branche créée)

#### Step 1: Converter offline `cmd/pak-titles`
**Description:** programme pur Go : parse `AccountTitleData.xml` (chaque `AccountTitleData` →
`id`, clé `Name`, `FontColor`, `StatType1..3`/`StatValue1..3`) et résout la clé dans
`StringData.xml` → `Fr`/`En`. Émet `internal/domain/data/titles.json` :
`{ "source": "...", "titles": [ {"id":2100004,"nameFr":"Mains curieuses","nameEn":"Curious Hands",
"color":"BLUE","stats":[{"type":"Defence","value":10}]}, … ] }`. En-tête doc = d'où viennent les
données + comment régénérer (comme `cmd/pak-catalog`). Défauts d'entrée = `tmp/pak/*.xml`.
**Verification:** `go run ./cmd/pak-titles` produit 108 titres ; contrôle ponctuel `2100004` →
« Mains curieuses / Curious Hands », toutes les catégories `id/64 ∈ {32812,32828,32843,32844,
32859,32875,32876}`. `just ci`.
**Commit:** `feat(cmd): pak-titles converter → titles.json`

#### Step 2: Titres — lecture (domain)
**Description:** ajouter `data/titles.json` à `//go:embed` (`catalog.go`). Nouveau `title.go` :
`init()` charge le seed ; type `Title{ ID int64; Category, Bit int; NameFR, NameEN, Color string;
Stats []TitleStat; Unlocked bool }` ; `Titles()` lit `tb_title (CATEGORY, BIT_FIELD)` en
`map[int]uint64`, dérive `(cat,bit)=titlePos(id)` et `Unlocked = bits[cat]>>bit&1==1`. Ordre =
ordre du catalogue (croissant par id).
**Verification:** `DSA_SAVE=… just test` : 108 titres, exactement 6 `Unlocked` (2100004, 2101000,
2104007, 2104012, 2104013, 2104100). `just ci`.
**Commit:** `feat(domain): read titles with unlocked state from tb_title`

#### Step 3: Titres — écriture (domain)
**Description:** `SetTitleUnlocked(id int64, unlocked bool)` = read-modify-write du bit dans la
catégorie (`INSERT OR REPLACE tb_title (USER_DBID,CATEGORY,BIT_FIELD)`, préserve les autres bits ;
`FAV_BIT_FIELD` laissé à sa valeur / défaut 0). `UnlockAllTitles()` = OR des bits réels de chaque
catégorie (calculés depuis le catalogue), transactionnel — jumeau exact de `UnlockAllRecipes`.
Manipulation en `uint64` → `int64(field)` pour la catégorie 32843 (bits hauts).
**Verification:** round-trip vrai save : cocher un titre absent → bit posé, `Titles()` le voit ;
décocher → bit retiré ; `UnlockAllTitles` → les 108 `Unlocked`, aucun bit hors des 7 masques ;
ré-ouverture du save OK. `just ci`.
**Commit:** `feat(domain): unlock titles (per-title and unlock-all)`

#### Step 4: Endpoints web
**Description:** `GET /api/game/titles` → `{titles: [...]}` ; `POST /api/game/titles/unlock
{id, unlocked}` ; `POST /api/game/titles/unlock-all`. `needSave` + `writeJSON`, mêmes conventions
que `handleRecipes`/`handleSetRecipeKnown`/`handleUnlockRecipes`. Enregistrer dans `web.go`.
**Verification:** `just ci` (save read-only tant que l'utilisateur n'a pas cliqué « Enregistrer » →
écriture atomique + backup déjà en place).
**Commit:** `feat(web): titles endpoints`

#### Step 5: Panneau UI
**Description:** écran « Titres » sur le patron du panneau Recettes : liste des 108 titres (nom Fr/En
selon la langue, pastille de couleur `FontColor`, bonus de stats en libellé secondaire), case à
cocher « débloqué » par titre, bouton « Tout débloquer ». Optionnel : groupement/filtre par couleur.
Réutilise les composants existants (liste, toggles, bouton d'action de masse).
**Verification:** revue visuelle + `manual_tests.md` renseigné.
**Commit:** `feat(web): titles panel`

#### Step 6: Docs
**Description:** `docs/database.md` (schéma `tb_title` + sémantique bitmask + `FAV_BIT_FIELD` noté
hors scope) ; `docs/content-ids.md` (titres `2100000`–`2104xxx`, 7 catégories) ; `docs/switches.md`
(note : `tb_title` partage la structure `(CATEGORY, BIT_FIELD)`) ; `DESIGN.md` / `README.md` (écran
Titres) ; `tmp/pak/README.md` (ajouter `AccountTitleData.xml` à la recette d'extraction).
**Verification:** relecture ; `just ci`.
**Commit:** `docs: document tb_title and the titles screen`

> Note : Steps 4→6 peuvent être fusionnés en commits couplant code+doc si le découpage paraît trop
> fin à l'implémentation ; chaque commit doit passer les gates et emporter la doc du code qu'il décrit.

---

### Décisions techniques
- **Déblocage seulement** : la case pilote un bit ; `FAV_BIT_FIELD` (titre affiché) reste hors
  scope tant que sa sémantique n'est pas établie par diff de deux vrais saves.
- **Bits réels uniquement** : `UnlockAllTitles` n'allume que les 108 bits du catalogue, jamais un
  masque plein — évite de poser des bits sans titre correspondant.
- **`uint64` en interne, `int64` en base** : gère la catégorie 32843 (bits > 62) par complément à
  deux, comme les masques `tb_switch` des recettes.
- **Catalogue data-driven, offline** : `cmd/pak-titles` reste un outil externe (fixtures `tmp/`,
  copyright jeu) ; l'éditeur reste **pur Go**. On régénère `titles.json` seulement quand le jeu
  patche ses données.
- **Pas d'icônes de titre** : aucune correspondance id→sprite confirmée ; texte + pastille couleur.

### Quality gates
- [ ] `just ci` passe (fmt-check + vet + lint + test + build)
- [ ] Round-trip vrai save vert (`DSA_SAVE=… just test`) : lecture 6 débloqués, unlock/unlock-all,
      réouverture
- [ ] Docs synchronisées (même commit que le code)
- [ ] Commits atomiques sur `feat/titles`, jamais sur `main`
- [ ] Validation manuelle en jeu (voir `manual_tests.md`) avant merge par l'utilisateur
