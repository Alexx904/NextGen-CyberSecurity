# User Stories

> Responsabile: Business Analyst. Storie derivate dai requisiti esistenti, non nuove promesse di prodotto.

| ID | Storia | Riferimenti e accettazione |
|---|---|---|
| US-01 | Come operatore voglio selezionare un'azienda per lavorare sui suoi asset. | FR-09/FR-15; UC-07/UC-14; verifica scope sulle viste coinvolte. |
| US-02 | Come operatore voglio scoprire dispositivi e importarli tra i target. | FR-10; UC-08; verifica risultati, importazione, errore e lista vuota. |
| US-03 | Come SecOps voglio avviare e seguire scansioni con parametri e consenso adeguati. | FR-02/FR-07/FR-12/FR-16; UC-02/UC-11/UC-15; verifica consenso invasivo, avanzamento e annullamento. |
| US-04 | Come analista voglio esaminare vulnerabilità e percorsi di attacco. | FR-04/FR-14/FR-17; UC-04/UC-13; verifica dettaglio e assenza di risultati. |
| US-05 | Come operatore voglio gestire firewall remoti con possibilità di ripristino. | FR-11; UC-09/UC-10; verifica autorizzazioni e safety rollback. |
| US-06 | Come auditor voglio esportare report leggibili per il cliente. | FR-06/FR-08/FR-18; UC-06/UC-17; verifica coerenza dati e layout. |
| US-07 | Come SecOps voglio neutralizzare un chokepoint con controllo del rollback. | FR-19; UC-18; pianificata in Epic 28. |
| US-08 | Come auditor voglio valutare gap di conformità per azienda. | FR-20; UC-19; pianificata in Epic 29. |
| US-09 | Come operatore voglio playbook di remediation assistiti. | FR-21; UC-20; pianificata in Epic 30. |

Il dettaglio normativo resta in [REQUIREMENTS](REQUIREMENTS.md) e [USE_CASES](USE_CASES.md). Lo stato effettivo e l'accettazione sono in [TASK_LIST](../04-development/TASK_LIST.md). Nuove storie richiedono criteri osservabili happy/error/edge e collegamenti ai test.
