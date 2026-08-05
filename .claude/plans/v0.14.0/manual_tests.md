# Tests manuels — v0.14.0 (polish thème, T1–T6)

Revue visuelle, à faire dans le navigateur sur une save ouverte (le rendu UI
n'est pas automatisable — cf. `PROCEDURE_PLANS.md` §4). Cocher au fil du dev.

## T1 — Spinner natif retiré
- [ ] Monnaies : le champ du stepper n'affiche **plus** de flèches haut/bas.
- [ ] Inventaire → détail : quantité en stock, plus de flèches.
- [ ] Inventaire → « Remplir » : champ sans flèches.
- [ ] Équipement : champs Enchant / XP sans flèches.
- [ ] Aucun résidu de spinner sur Chromium **et** Firefox.

## T2 — Modale maison (remplace prompt/confirm)
- [ ] Renommer un objet (crayon ✎) ouvre une **modale thémée**, pas un popup OS.
      Valider avec Entrée écrit le nom ; Échap / clic overlay annule (aucun effet).
- [ ] Cuisine → « Tout débloquer » : modale de confirmation (variante danger),
      Annuler = rien, Confirmer = recettes marquées connues.
- [ ] Inventaire → « Remplir » : modale de confirmation, mêmes comportements.
- [ ] Clavier : focus initial correct (champ pour prompt, bouton primaire pour
      confirm) ; Tab reste dans la modale ; Échap ferme.
- [ ] `grep -n "window.prompt\|window.confirm\|[^.]confirm(\|[^.]prompt(" app.js`
      ne renvoie plus d'appel natif.

## T3 — Toggle « verrou »
- [ ] Équipement + Gemmes : le verrou est un **interrupteur thémé**, plus une
      case à cocher OS. État coché = accent. Bascule = même effet qu'avant
      (round-trip via « Écrire »). Focusable au clavier, activable par Espace.

## T4 — Focus unifié
- [ ] Tab à travers : dossier (home), stepper, Enchant/XP, « Remplir »,
      recherche cuisine, filtre SQL → **même** anneau de focus discret partout,
      pas d'outline bleu navigateur résiduel.
- [ ] Au clic souris, pas d'anneau intempestif (`:focus-visible`).

## T5 — Bordures de champs
- [ ] Tous les champs éditables partagent le même langage visuel (bordure au
      repos cohérente, accent au focus). Plus de mélange accent/gris arbitraire.

## T6 — Scrollbars thémées
- [ ] Inventaire (rail + grille + détail), SQL (rail + grille), Cuisine (grille +
      détail) : scrollbars sombres discrètes, cohérentes avec le fond, sur
      Chromium et Firefox. Défilement molette/clavier intact.

## Non-régression générale
- [ ] Les 9 écrans s'affichent sans erreur console (home, inventaire, monnaies,
      personnages, équipe, équipement, gemmes, cuisine, SQL brut).
- [ ] Bascule FR/EN toujours fonctionnelle.
- [ ] « Écrire dans la save » fonctionne, badge « N modif. » se remet à zéro,
      toast + backup `.bak` inchangés.
