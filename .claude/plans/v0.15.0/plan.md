## Plan: Costumes & Familiers (montures) — deux nouveaux écrans joueur

**Type:** edit capability + UI (domain → web → docs), sur le patron Cuisine/Équipement/Gemmes
**Objective:** Deux panneaux « Costumes » et « Familiers » qui listent le possédé résolu
(nom/icône via le catalogue), permettent de **débloquer** un cosmétique absent (INSERT) et
de l'**équiper / déséquiper** sur un personnage. Tables aujourd'hui accessibles seulement
en SQL brut.
**Why:** BACKLOG P1 (`Costumes / Apparences`, `Montures / Véhicules`), risque faible
(cosmétique, aucun impact stat/progression). Le modèle a été établi par diff de deux vrais
saves (voir « Modèle confirmé » ci-dessous).
**Layer(s):** domain, web, docs (aucun changement sqlcipher / save / pak / oodle)

---

### Contexte — modèle confirmé (diff de deux vrais saves, 2026-08-05)

Établi en comparant un save avant/après avoir équipé « un costume + une arme » sur Cerese
(`CHARACTER_CID 10029`).

**Costumes — `tb_costume`** `(COSTUME_DBID pk, USER_DBID, COSTUME_CID, CREATED_DATE, PARTS_ON, IS_NEW)`
- Possession = une ligne par cosmétique possédé. Catalogue : plage CID `9990008`–`9990031`
  (16 nommés, `category=costume`, tous `type=COSTUME`).
- **Équiper = poser le `CHARACTER_CID` du perso dans `EQUIP_CHARACTER_CID`** ; `0` = non porté.
- **Non exclusif** : un perso porte **plusieurs lignes `tb_costume` simultanément**.
  Fait observé : équiper « une tenue + un skin d'arme » a mis DEUX lignes (9990008 + 9990009)
  toutes deux à `10029`.
- L'espace `999xxxx` mélange **tenues ET skins d'arme** — tous `type=COSTUME`, indistinguables
  par le type. Ils vont par **paires pair/impair** (8+9, 24+25 : tenue + skin d'arme assorti).
  Hypothèse d'appairage **non vérifiée** → on n'enforce rien, on affiche à plat.
- `PARTS_ON` reste `1` même équipé → masque de parties visibles, **pas** le flag d'activation.
  `IS_NEW` = badge « nouveau ». `tb_character` n'a **aucune** colonne d'apparence : l'état
  d'équipement vit uniquement dans `tb_costume`.

**Familiers = montures — `tb_vehicle`** `(VEHICLE_DBID pk, USER_DBID, VEHICLE_CID, CREATED_DATE)`
- « Familiers » = `tb_vehicle` (créatures montables : loups célestes, griffons, dragons,
  chats ailés), **pas** `tb_summon` (table vide / compteurs obscurs, hors scope).
- Catalogue : plage CID `1320000`–`1320033` (30 nommés, `category=mount`). `1320033` non
  traduit (coréen `흉악한 새끼 용`) → trou de données catalogue, à signaler pas à masquer.
- **Équiper = `tb_equip_mount.VEHICLE ← VEHICLE_DBID` possédé** (une monture par perso,
  une ligne `(USER_DBID, CHARACTER_CID)` par perso ; `VEHICLE=0` = aucune). Un perso peut
  ne pas avoir de ligne `tb_equip_mount` → l'équiper nécessite alors un INSERT de la ligne.

**Débloquer (les deux)** = INSERT avec un `*_DBID` unique fraîchement forgé. Les DBID observés
tiennent dans un uint32 (~26 M–4,04 Md), d'allure aléatoire, distincts de l'espace 64-bit des
instances d'équipement. → forge aléatoire + vérif d'absence.

---

### Scope

**In:**
- **domain** : accès typés `Costumes()` / `Vehicles()` (possédé résolu via catalogue) ;
  énumération catalogue avec drapeau *possédé* (pour l'écran de déblocage) ; `Characters()`
  déjà présent réutilisé pour la cible d'équipement ; par-personnage la monture équipée
  (`tb_equip_mount`) et les costumes portés.
- **domain écriture** : `UnlockCostume(cid)` / `UnlockVehicle(cid)` (INSERT, no-op si déjà
  possédé) ; `SetCostumeEquip(dbid, characterCID)` (0 = déséquiper) ; `SetMount(characterCID,
  vehicleDBID)` (0 = aucune ; INSERT de la ligne `tb_equip_mount` si absente). Forge de DBID
  unique mutualisée.
- **web** : deux panneaux + endpoints `/api/game/costumes*` et `/api/game/familiers*`.
- **docs** : `docs/database.md` (schémas + sémantique d'équipement des 3 tables),
  `docs/content-ids.md` (plages déjà présentes, préciser tenue+skin), `DESIGN.md` / `README.md`
  (liste des écrans), `CHANGELOG` via git-cliff.

**Out:**
- `tb_summon` / compagnons / mercenaires (chantier distinct).
- **Suppression** d'un cosmétique possédé (retrait de ligne) — risqué si équipé, faible valeur.
  Déblocage uniquement ; documenté comme non couvert.
- Distinction tenue vs skin d'arme dans l'UI (appairage non vérifié) — affichage à plat.
- Édition de `PARTS_ON` (sémantique non figée), de `IS_NEW`, ou des accessoires de monture
  (`tb_equip_mount.ACC_*`, `TALISMAN_*`, `KARMA`).
- Traduction du CID `1320033` (donnée pak, hors de ce chantier).
- Tout changement `sqlcipher` / `save` / `pak` / `oodle`.

---

### Files touched
- [ ] `internal/domain/costume.go` — types + `Costumes()`, `CostumeCatalog()`, `UnlockCostume`, `SetCostumeEquip`
- [ ] `internal/domain/vehicle.go` — types + `Vehicles()`, `VehicleCatalog()`, `Mounts()`, `UnlockVehicle`, `SetMount`
- [ ] `internal/domain/dbid.go` — `mintDBID(table, pkCol)` forge d'ID unique (crypto/rand, pur Go)
- [ ] `internal/domain/*_test.go` — accès + round-trip unlock/equip sur vrai save (`DSA_SAVE`)
- [ ] `internal/web/game.go` — handlers costumes + familiers
- [ ] `internal/web/web.go` — enregistrement des routes
- [ ] `internal/web/static/…` — deux panneaux UI (liste possédé, catalogue à débloquer, équiper)
- [ ] `docs/database.md`, `docs/content-ids.md`, `DESIGN.md`, `README.md`
- [ ] `.claude/plans/v0.15.0/manual_tests.md`

---

### Atomic steps

#### Step 0: Audit Phase 0
**Description:** `nix develop --command just ci` sur `main` propre, résultat consigné dans
`phase0_results.md`. Créer la branche `feat/costumes-familiers`.
**Verification:** `just ci` vert avant toute ligne de code.
**Commit:** — (pas de commit ; branche créée)

#### Step 1: Forge de DBID unique (domain, pur Go)
**Description:** `mintDBID(table, pkCol string) (int64, error)` : tire un uint32 aléatoire
(`crypto/rand`), boucle tant que le PK existe déjà dans la table. Aucune dépendance CGO.
**Verification:** test unitaire (unicité vs un set simulé) + `just ci`.
**Commit:** `feat(domain): mint unique instance DBIDs for inserts`

#### Step 2: Costumes — lecture (domain)
**Description:** type `Costume{ Item; DBID string; EquipCharacterCID; PartsOn; IsNew }` ;
`Costumes()` (possédé résolu `LookupCtx(cid,"costume")`, perso équipé résolu si ≠0) ;
`CostumeCatalog()` = toutes les entrées `category=costume` + drapeau *possédé*.
**Verification:** `DSA_SAVE=… just test` (10 possédés, 6 débloquables, aucun équipé sur le
fixture) + `just ci`.
**Commit:** `feat(domain): read owned costumes + costume catalog`

#### Step 3: Costumes — écriture (domain)
**Description:** `UnlockCostume(cid)` (INSERT `(mintDBID, uid, cid, now, 1, 1)`, no-op si déjà
possédé) ; `SetCostumeEquip(dbid, characterCID)` (UPDATE `EQUIP_CHARACTER_CID`, `0`=déséquiper).
Non exclusif : aucune contrainte d'unicité par perso.
**Verification:** round-trip sur vrai save : unlock → présent ; equip → `EQUIP_CHARACTER_CID`
posé ; unequip → `0`. Ré-ouverture OK. `just ci`.
**Commit:** `feat(domain): unlock and equip costumes`

#### Step 4: Familiers — lecture (domain)
**Description:** type `Vehicle{ Item; DBID string }` ; `Vehicles()` (résolu `"mount"`) ;
`VehicleCatalog()` (drapeau possédé) ; `Mounts()` = par perso (`CHARACTER_CID`) la monture
équipée résolue depuis `tb_equip_mount.VEHICLE` → `tb_vehicle.VEHICLE_CID`.
**Verification:** `DSA_SAVE=… just test` (montures possédées, mapping par perso, Eileen sans
monture) + `just ci`.
**Commit:** `feat(domain): read owned vehicles + per-character mounts`

#### Step 5: Familiers — écriture (domain)
**Description:** `UnlockVehicle(cid)` (INSERT, no-op si possédé) ; `SetMount(characterCID,
vehicleDBID)` : UPDATE `tb_equip_mount.VEHICLE` ; si aucune ligne pour le perso, INSERT une
ligne minimale `(uid, characterCID, VEHICLE=…)`. `vehicleDBID=0` = aucune monture.
**Verification:** round-trip : unlock → présent ; équiper un perso qui a déjà une ligne, puis
un perso sans ligne (INSERT) ; ré-ouverture OK. `just ci`.
**Commit:** `feat(domain): unlock and equip vehicles (familiers)`

#### Step 6: Endpoints web
**Description:** `GET /api/game/costumes` (possédé + catalogue + personnages) ;
`POST /api/game/costumes/unlock {cid}` ; `POST /api/game/costumes/equip {dbid, characterCid}` ;
mêmes trois pour `/api/game/familiers*`. Wrap `needSave`, `writeJSON`, mêmes conventions que
`handleGems`/`handleRecipes`.
**Verification:** `just ci` (le save reste read-only tant que l'utilisateur ne clique pas
« Enregistrer » → écriture atomique + backup déjà en place).
**Commit:** `feat(web): costumes & familiers endpoints`

#### Step 7: Panneaux UI
**Description:** deux écrans jumeaux : liste du possédé (icône, nom, badge « nouveau »,
perso équipé + sélecteur pour (dés)équiper), section « Débloquer » listant le catalogue non
possédé avec bouton. Familiers idem avec « monture par personnage ». Réutilise les composants
existants (grille d'items, sélecteur de perso, icônes via `/api/icon`).
**Verification:** revue visuelle + `manual_tests.md` renseigné.
**Commit:** `feat(web): costumes & familiers panels`

#### Step 8: Docs
**Description:** `docs/database.md` (schémas `tb_costume` / `tb_vehicle` / `tb_equip_mount`
+ sémantique d'équipement non exclusive + tenue/skin regroupés) ; `docs/content-ids.md`
(préciser `999xxxx` = tenues+skins, `1320033` non traduit) ; `DESIGN.md` / `README.md` (deux
écrans de plus) ; note sur `tb_summon` laissé hors scope.
**Verification:** relecture ; `just ci`.
**Commit:** `docs: document costume & vehicle tables and screens`

> Note : Steps 6→8 peuvent être fusionnés en commits couplant code+doc si le découpage
> paraît trop fin à l'implémentation ; l'essentiel est que chaque commit passe les gates et
> que la doc parte avec le code qu'elle décrit.

---

### Décisions techniques
- **Non-exclusivité assumée** : on n'impose pas « un costume par perso » — c'est le
  comportement observé du jeu (tenue + skin d'arme sur le même perso).
- **Pas d'appairage tenue/skin** dans l'UI (hypothèse non vérifiée) : affichage à plat, le
  joueur équipe/déséquipe chaque cosmétique indépendamment.
- **Déblocage seulement, pas de suppression** : retirer un cosmétique possédé est risqué
  (s'il est équipé) et de faible valeur ; explicitement hors scope.
- **Forge de DBID** : uint32 aléatoire + vérif d'absence, mutualisée entre costumes/véhicules.
- **`SetMount` gère l'absence de ligne** `tb_equip_mount` par un INSERT (certains persos n'en
  ont pas), là où costumes est toujours un UPDATE d'une ligne existante ou un INSERT global.

### Quality gates
- [ ] `just ci` passe (fmt-check + vet + lint + test + build)
- [ ] Round-trip vrai save vert (`DSA_SAVE=… just test`) : unlock + equip + réouverture
- [ ] Docs synchronisées (même commit que le code)
- [ ] Commits atomiques sur `feat/costumes-familiers`, jamais sur `main`
- [ ] Validation manuelle en jeu (voir `manual_tests.md`) avant merge par l'utilisateur
