# Dokumentasi PMO

Dokumen dikelompokkan berdasarkan fungsinya. Jalankan perintah skrip dari root proyek PMO.

## Kebutuhan dan arsitektur

- [ARCHITECTURE.md](<architecture/ARCHITECTURE.md>)
- [ERD-CONCEPTUAL.md](<architecture/ERD-CONCEPTUAL.md>)
- [PERMISSION-MATRIX.md](<architecture/PERMISSION-MATRIX.md>)
- [SRS.md](<architecture/SRS.md>)

## Perencanaan dan status implementasi

- [DEVELOPMENT-BACKLOG.md](<planning/DEVELOPMENT-BACKLOG.md>)
- [IMPLEMENTATION-GAP-ANALYSIS.md](<planning/IMPLEMENTATION-GAP-ANALYSIS.md>)
- [PHASED-IMPLEMENTATION-PLAN.md](<planning/PHASED-IMPLEMENTATION-PLAN.md>)
- [PMO-CONTROL-TOWER-ANALYSIS.md](<planning/PMO-CONTROL-TOWER-ANALYSIS.md>)

## UAT dan rilis

- [RELEASE-CANDIDATE-RUNBOOK.md](<uat/RELEASE-CANDIDATE-RUNBOOK.md>)
- [UAT-BUSINESS-FLOW-SCENARIOS.md](<uat/UAT-BUSINESS-FLOW-SCENARIOS.md>)
- [UAT-DEMO-GUIDE.md](<uat/UAT-DEMO-GUIDE.md>)
- [UAT-DEPLOYMENT-GUIDE.md](<uat/UAT-DEPLOYMENT-GUIDE.md>)
- [UAT-FINAL-CHECKLIST.md](<uat/UAT-FINAL-CHECKLIST.md>)

## Bahan referensi

- [Deck Konsep PMO SDA edit rev.pdf](<references/Deck Konsep PMO SDA edit rev.pdf>)
- [Deck PMO SDA.pptx](<references/Deck PMO SDA.pptx>)
- [WhatsApp Image 2026-08-20 at 22.40.43.jpeg](<references/WhatsApp Image 2026-08-20 at 22.40.43.jpeg>)

## Skrip pengujian

Skrip smoke dan regresi berada di [`scripts/smoke/`](../scripts/smoke/). Skrip deployment, backup, dan pengelolaan UAT lokal berada di [`scripts/`](../scripts/).

Contoh menjalankan regression gate setelah layanan dan data pengujian siap:

```bash
bash scripts/smoke/smoke-test-rel-001-regression.sh
```
