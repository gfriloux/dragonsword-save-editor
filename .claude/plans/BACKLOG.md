# Backlog — pistes d'amélioration

> Liste vivante des TODO envisagés, **hors** plans formels `vX.Y.Z/`. Chaque entrée
> devient un plan sous `.claude/plans/` avant tout code (voir `PROCEDURE_PLANS.md`).
> Priorités : **P1** (forte valeur / à faire tôt) · **P2** · **P3** (confort).
> État au dernier shipping : **v0.14.0**.

## A. Nouveaux écrans joueur (P1)

Tables présentes dans le save mais accessibles seulement via « SQL brut ». Même
patron que Cuisine/Équipement : accesseur typé `internal/domain` → endpoint
`/api/game/*` → panneau UI.

- ~~**Entrepôt / Coffre** — `tb_storage`, `tb_storage_bm_shop`.~~ **ABANDONNÉ — feature
  coupée du jeu (vérifié 2026-08-05).** Tables vestigiales : jamais peuplées (aucun
  row count dans un vrai save), aucune string UI dans les paks (« Coffre » = treasure
  boxes seulement), aucun flag « stockable » sur les objets, et introuvable en jeu.
  Les guides web mentionnant un « Storage Chest » sont du contenu IA générique
  contredit par les strings shippées. Éditer cette table ne mappe à rien de vivant.
- ~~**Costumes / Apparences** — `tb_costume`.~~ **SHIPPED v0.15.0** — écran Costumes
  (débloquer + équiper par perso, non exclusif : tenue + skin d'arme).
- [ ] **Titres** — `tb_title`. Déblocage. **PLANIFIÉ v0.16.0** (`.claude/plans/v0.16.0/`) —
  étude faite 2026-08-06 : bitmask comme les recettes, catalogue data-driven (108 titres,
  `AccountTitleData.xml`).
- ~~**Montures / Véhicules** — `tb_vehicle`, `tb_equip_mount`.~~ **SHIPPED v0.15.0** —
  écran Familiers (débloquer + une monture par perso via `tb_equip_mount`).
- [ ] **Compétences** — `tb_skill_growth`. Niveaux de skill par personnage.
- [ ] **Compagnons / Mercenaires** — `tb_summon`, `tb_mercenary_`, `tb_mercenary_trainee_`.
- [ ] **Carte & voyage rapide** — `tb_field_statue`, `tb_field_statue_charge`, `tb_map_memo`.
- [ ] **Journal de quêtes** — `tb_completed_quest`, `tb_quest_hold`, `tb_quest_complete`.
- [ ] **Donjons / Tour d'épreuve** — `tb_trial_tower`, `tb_crack_dungeon`, `tb_active_dungeon`.
- [ ] **Succès** — `tb_achievement_category_step`, `tb_achievement_count`.

Ordre valeur/risque suggéré : Costumes/Titres → Compétences → Montures.
(L'Entrepôt, ci-dessus, sortait en tête mais s'est révélé non viable — voir la note.)

## B. Rendre éditable ce qui est en lecture seule (P2)

- [ ] **Personnages** — édition niveau/exp (avec bornes). Aujourd'hui lecture seule.
- [ ] **Équipe** — recomposition des escouades, encadrée par de la validation pour
  ne pas produire un état que le jeu refuse.

## C. UX & sécurité d'édition (P1–P2)

- [ ] **Revue globale avant écriture** (P1) — récapitulatif de *toutes* les modifs en
  attente, tous écrans confondus, dans une modale « Vérifier les changements »
  (au-delà du diff par cellule déjà présent dans SQL brut).
- [ ] **Gestion des backups depuis l'UI** (P1) — lister et **restaurer** un `.bak`
  (le backup existe déjà côté `internal/save`, mais pas de restauration exposée).
- [ ] **Détection du jeu ouvert** (P1) — refuser/avertir si la DB est verrouillée
  (l'invariant « jamais écrire pendant que le jeu tourne » n'est aujourd'hui qu'un
  texte d'avertissement).
- [ ] **Validation / bornes** (P2) — max monnaie, max stack… pour éviter les valeurs
  rejetées in-game.
- [ ] **Recherche transverse** (P2) dans le navigateur SQL brut.
- [ ] **Export JSON** (P3) d'un save en lecture seule pour inspection/partage (sans art).

## D. Profondeur des données (P2–P3)

- [ ] **Recettes spéciales** (P2) — boissons glacées `1999xxx`, `1423001`, `1430920`,
  hors `CookRecipeData.xml` (table différente), à intégrer dans Cuisine.
- [ ] **Stats & descriptions d'objets** (P3) depuis les paks → infobulles enrichies.
- [ ] **Locales supplémentaires** (P3) si les paks en contiennent d'autres que FR/EN.

## E. Robustesse & ingénierie (P2–P3)

- [ ] **Test round-trip golden** (P2) sur un vrai save via `DSA_SAVE`, intégré au `just`.
- [ ] **Fuzz** (P2) sur `internal/pak` et `internal/sqlcipher`.
- [ ] **Détection de dérive de schéma** (P2) entre patchs du jeu — message clair
  plutôt qu'un crash si une colonne bouge.
- [ ] **Décodeur Kraken/Oodle pur-Go natif** (P3) pour supprimer la dépendance
  WASM/wazero. Confort interne : l'invariant CGO-free est déjà tenu.

## F. Distribution (P3)

- [ ] **Clarifier le workflow de release** — le tag `v*` déclenche un workflow, mais
  `DESIGN.md` dit « pas de CI GitHub » : trancher la contradiction.
- [ ] **Packaging du binaire Windows** (signature / archive de release).
