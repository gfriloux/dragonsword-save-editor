# Tests manuels — Titres (v0.16.0)

> Renseignés au fil du dev, joués à la validation. Le jeu doit être **fermé** avant toute
> écriture. L'éditeur ouvre en lecture seule, sauvegarde atomique + backup automatique.
> Tester sur une **copie** du vrai save, jamais l'original.

## Préparation
- [ ] Copier le save live vers un emplacement de test, pointer l'éditeur dessus.
- [ ] Faire un backup manuel à côté (ceinture + bretelles).

## Titres — lecture
- [ ] Le panneau liste les 108 titres avec nom FR (ou EN selon la langue) et pastille de
      couleur (BLUE/BROWN/GREEN/RED/VIOLET).
- [ ] Exactement 6 titres apparaissent « débloqués » sur le fixture : 2100004 (« Mains
      curieuses »), 2101000, 2104007, 2104012, 2104013, 2104100.
- [ ] Le bonus de stats s'affiche pour les titres qui en accordent un (ex. « Mains curieuses »
      → Défense +10).

## Titres — écriture (jeu fermé)
- [ ] Cocher un titre absent → il passe « débloqué » après enregistrement.
- [ ] Ouvrir le jeu : le titre débloqué est présent et sélectionnable dans l'écran des titres.
- [ ] Décocher un titre → il repasse « non débloqué » ; en jeu il disparaît de la liste.
- [ ] « Tout débloquer » → les 108 titres passent débloqués ; en jeu, tous présents.
- [ ] Cas bits hauts : vérifier qu'un titre de la catégorie 32843 (ID ~2101952+) se débloque
      correctement (pas d'erreur de signe / masque).

## Sécurité / non-régression
- [ ] Après chaque écriture, un `.bak` horodaté est créé.
- [ ] Le save réédité se ré-ouvre sans erreur dans l'éditeur.
- [ ] Le jeu charge le slot sans message d'erreur ni reset ; les autres données (personnages,
      inventaire) intactes.
- [ ] Aucun autre écran (Cuisine, Équipement, Gemmes, Costumes, Familiers…) n'est cassé.
