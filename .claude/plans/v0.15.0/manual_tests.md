# Tests manuels — Costumes & Familiers (v0.15.0)

> Renseignés au fil du dev, joués à la validation. Le jeu doit être **fermé** avant toute
> écriture. L'éditeur ouvre en lecture seule, sauvegarde atomique + backup automatique.
> Tester sur une **copie** du vrai save, jamais l'original.

## Préparation
- [ ] Copier le save live vers un emplacement de test, pointer l'éditeur dessus.
- [ ] Faire un backup manuel à côté (ceinture + bretelles).

## Costumes — lecture
- [ ] Le panneau liste les 10 costumes possédés avec nom FR + icône.
- [ ] Les costumes déjà équipés en jeu s'affichent avec le bon personnage (ex. avant test :
      Cerese porte « Le serment d'Ophelia » + « Le vœu de silence » ; Dana porte
      « Serveuse aux framboises » + « Service sucré »).
- [ ] La section « Débloquer » liste les 6 costumes non possédés (dont « Habit du Grand
      Esprit », « Pétales à la crête des vagues »).

## Costumes — écriture (jeu fermé)
- [ ] Débloquer un costume absent → il apparaît dans le possédé après enregistrement.
- [ ] Ouvrir le jeu : le costume débloqué est présent et équipable en jeu.
- [ ] Équiper un costume sur un perso via l'éditeur → en jeu, le perso le porte.
- [ ] Déséquiper (remettre à « aucun ») → en jeu, le costume n'est plus porté.
- [ ] Équiper une **tenue + un skin d'arme** sur le même perso → les deux tiennent (non
      exclusif), le jeu affiche bien les deux.

## Familiers (montures) — lecture
- [ ] Le panneau liste les montures possédées (nom FR + icône).
- [ ] La monture équipée par personnage est correcte (ex. Lute → « Loup géant des Abysses »,
      Johnny → « Jeune dragon courageux » ; Eileen → aucune).
- [ ] « Débloquer » liste les montures manquantes ; `1320033` s'affiche sans planter malgré
      son nom non traduit.

## Familiers — écriture (jeu fermé)
- [ ] Débloquer une monture absente → présente après enregistrement, montable en jeu.
- [ ] Équiper une monture sur un perso qui en avait déjà une → remplacée en jeu.
- [ ] Équiper une monture sur un perso **sans ligne** `tb_equip_mount` (ex. si un perso
      neuf n'en a pas) → la ligne est créée, la monture apparaît en jeu.
- [ ] Retirer la monture (« aucune ») → en jeu, plus de monture.

## Sécurité / non-régression
- [ ] Après chaque écriture, un `.bak` horodaté est créé.
- [ ] Le save réédité se ré-ouvre sans erreur dans l'éditeur.
- [ ] Le jeu charge le slot sans message d'erreur ni reset.
- [ ] Aucun autre écran (Cuisine, Équipement, Gemmes…) n'est cassé.
