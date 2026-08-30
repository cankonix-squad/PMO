# PMO National Project Control Tower Analysis
# PMO - Direktorat Jenderal Sumber Daya Air

**Versi**: 0.2.2  
**Tanggal Analisis**: 2026-08-21  
**Status**: Baseline kebutuhan dan roadmap; bukan bukti implementasi

---

## 1. Tujuan Dokumen

Dokumen ini menerjemahkan materi rencana pengembangan terbaru menjadi baseline kebutuhan PMO sebagai **PMO National Project Control Tower**. Dokumen ini menjadi penghubung antara materi bisnis, SRS, arsitektur, ERD, permission matrix, gap analysis, dan development backlog.

Strategi produk yang disepakati:

- PMO adalah sistem operasional, workflow, audit, dan sumber data utama.
- Power BI menjadi dashboard eksekutif tahap awal melalui dataset terkontrol.
- Dashboard native PMO tetap digunakan untuk pekerjaan operasional dan dikembangkan bertahap.
- Informasi disajikan dalam tiga level: nasional, program/sektor, dan detail proyek.

### 1.1 Current UI Translation Status (2026-08-21)

- Login PMO dan native executive dashboard telah diterjemahkan dari mockup dengan shell Kementerian PU/Ditjen SDA, responsive KPI/panel, sidebar, dan data demo lokal.
- Menu dan route `/command-center` telah dibuat untuk memvalidasi information architecture: KPI, alert center, validation placeholder, action tracker derived, risk/issue summary, quality heuristic, map illustration, reporting schedule, dan watchlist.
- Implementasi visual awal bukan bukti capability selesai. Per 2026-08-26, P1-015 sudah `Done` untuk lifecycle API/UI dasar Command Center: validation SLA, corrective action, Health Score integration, persisted escalation, executive decision follow-up, permission, dan audit. GIS aktual, Level 1/2 resmi, Power BI/read model, dan advanced analytics tetap mengikuti roadmap P2/P3.
- Visual mockup harus diterjemahkan secara responsif dan profesional, bukan dipaksakan pixel-perfect pada seluruh aspect ratio. Stabilitas multi-viewport dan Next.js asset lifecycle ditangani oleh P1-016.

## 2. Inventaris dan Kedudukan Sumber

| Sumber | Kedudukan | Cara Menggunakan |
|--------|-----------|------------------|
| Gambar "Grand Design PMO National Project Control Tower Ditjen SDA" tanggal 2026-08-20 | Kebutuhan utama | Menetapkan struktur dashboard Level 1-3, indikator, sektor, GIS, evidence, dan roadmap integrasi |
| Slide Grand Design PMO SDA | Kebutuhan utama | Menetapkan visi, proses monitoring-corrective action-reporting, data input, dan output pimpinan |
| Slide Project Health Score | Kebutuhan utama konseptual | Menetapkan delapan dimensi score dan empat kelas health; bobot belum ditetapkan |
| Slide PMO SDA Platform Capability Stack | Kebutuhan utama | Menetapkan capability map dari command center sampai data integration |
| Mockup login, executive dashboard, dan PMO command center SDA | Referensi information architecture | Menentukan informasi dan workflow yang perlu tersedia, bukan spesifikasi pixel-perfect |
| Slide BP3R pada bagian akhir deck | Referensi pola PMO umum | Hanya pola generik seperti pipeline, bottleneck, quality assurance, dan executive insight; data/domain perumahan bukan requirement SDA |

Angka, nama proyek, tanggal, persentase, nilai kontrak, serta nama pejabat pada mockup bersifat ilustratif dan tidak boleh dipakai sebagai seed atau klaim data produksi.

## 3. Target Operating Model

### Level 1 - National Executive Control Tower

Pengguna utama adalah pimpinan nasional/Dirjen dan PMO pusat. Tampilan harus menyediakan total proyek, nilai kontrak/portofolio, progress fisik dan keuangan, deviasi waktu/biaya, distribusi health, peta nasional, trend, proyek kritis, keputusan yang dibutuhkan, dan manfaat nasional.

### Level 2 - Program/Sector Control

Pengguna utama adalah pengendali program/sektor. Agregasi minimal mencakup bendungan, irigasi, pengairan dan pengendalian banjir, air baku, dan pertanian. Pengguna dapat membandingkan program, melihat top deviasi waktu/biaya, risiko tinggi, realisasi anggaran, dan drill-down ke proyek.

### Level 3 - Project Control

Pengguna utama adalah PM, project controller, Satker/Balai/BBWS/BWS, dan tim lapangan. Detail proyek mencakup lokasi/DAS, kontrak, penyedia, konsultan, baseline, progress fisik/keuangan, deviasi, health explanation, jadwal versus realisasi, bukti lapangan, dokumen, isu, risiko, dan tindak lanjut.

## 4. Capability Map

| Capability | Target | Status PMO 2026-08-20 |
|------------|--------|----------------------------|
| Executive Command Center | Level 1 nasional, decision queue, critical ranking | Planned |
| Portfolio & Program Management | Program/sektor, wilayah, pipeline, target | Planned |
| Project Monitoring & Control | Project/task/milestone, progress, history | Basic operational MVP implemented |
| Issue, Risk & Corrective Action | Register, escalation, PIC, SLA, evidence | Planned; basic model sebagian tersedia |
| Contract & Financial Monitoring | Kontrak, penyedia, konsultan, realisasi, variance | Planned |
| Milestone & Schedule Control | Milestone, Gantt, baseline versus actual | Basic milestone implemented; Gantt/baseline planned |
| Field Progress & Evidence | Inspection, photo, geotag, timestamp | Planned |
| Document Management | Upload, version, download, archive | Planned |
| Reporting & Workflow | Mingguan/bulanan/triwulanan, approval, export | Planned |
| Analytics & Early Warning | Health score, alert rules, trend, ranking | Early warning dasar implemented; full analytics planned |
| Data Integration Platform | Import, source lineage, validation, sync | Planned |

## 5. Project Health Score

Project Health Score menggunakan delapan dimensi:

1. Schedule.
2. Physical progress.
3. Financial progress.
4. Contract.
5. Risk.
6. Issue.
7. Quality.
8. Procurement.

Keluaran diklasifikasikan menjadi `GREEN`, `YELLOW`, `RED`, atau `CRITICAL`. Setiap hasil harus menyimpan nilai komponen, formula version, data period, alasan klasifikasi, dan waktu kalkulasi agar dapat dijelaskan dan diaudit.

Materi belum menetapkan bobot, normalisasi, threshold numerik, atau aturan missing data. Karena itu:

- Formula dan threshold harus configurable serta versioned.
- Konfigurasi harus disetujui pemilik bisnis/PMO sebelum aktif.
- Sistem tidak boleh menganggap missing data sebagai nilai baik.
- Perubahan formula tidak boleh mengubah histori score lama.
- Mockup health gauge tidak menjadi persetujuan formula.

## 6. Data, Validasi, dan Freshness

Setiap angka eksekutif harus dapat ditelusuri ke proyek, periode, sumber data, dan status validasi. Data dari input manual, file, atau sistem eksternal masuk melalui validation queue sebelum dipakai untuk laporan resmi.

Minimum kontrol data:

- `organization_id` pada seluruh business data.
- Source, period, submitted by, submitted at, validated by, dan validated at.
- Status `DRAFT`, `SUBMITTED`, `VALID`, `REJECTED`, atau `STALE`.
- Completeness, freshness, rejection reason, dan SLA aging.
- Audit log untuk koreksi, validasi, perubahan formula, keputusan, dan export.
- Soft delete untuk business data; snapshot laporan/score yang sudah dipublikasikan tetap immutable.

## 7. Arsitektur Target Hybrid

```text
Input Manual / Field / External Systems
                 |
                 v
 PMO Operational API + PostgreSQL
                 |
       Validation & Rule Engine
                 |
        Analytics Read Model
          /               \
 Native Operational UI    Power BI Executive Dashboard
```

Power BI tidak menulis langsung ke tabel operasional. Dataset membaca view/read model yang tenant-safe, terdokumentasi, memiliki data dictionary, row-level security, refresh schedule, dan refresh monitoring.

## 8. Roadmap

| Tahap | Hasil Utama |
|-------|-------------|
| Hardening | Integritas soft delete, cleanup smoke, tenant guard |
| Data Foundation | Program/sektor/lokasi/DAS, kontrak, baseline, progress snapshot, validasi |
| PMO Core Process | Issue, risk, corrective action, field evidence, reporting |
| Project Intelligence | Gantt, variance, health score, project control dashboard |
| Control Tower | Level 2 program/sektor, Level 1 nasional, GIS, Power BI, decision support |
| Advanced Integration | Primavera P6, mobile inspection, BIM/digital twin, government/entity resolution, data governance; IoT dan AI/ML sebagai final/future capability |

IoT diparkir sebagai capability paling akhir: bentuk awalnya nanti adalah data gateway + device/source registry + telemetry validation + alert foundation, bukan MQTT/streaming platform penuh. AI/ML hanya dimulai setelah definisi indikator stabil, data tervalidasi, histori cukup, dan model dapat dijelaskan. Detail tiket dan dependency ada di [Development Backlog](./DEVELOPMENT-BACKLOG.md).

## 9. Current State versus Target State

- **Local operational MVP**: P0-001 sampai P0-014 sudah selesai dan terverifikasi lokal.
- **PMO core process**: `PMO-P1-001` Issue Management sudah selesai; fase aktif berikutnya dimulai dari `PMO-P1-002` Risk Management.
- **Full PMO Control Tower**: belum terpenuhi. Power BI, GIS, progress snapshot resmi, data validation, full health score, program dashboard, field inspection, dan integrasi eksternal masih roadmap.
- Status implementasi rinci harus selalu mengacu pada [Implementation Gap Analysis](./IMPLEMENTATION-GAP-ANALYSIS.md), sedangkan urutan delivery mengacu pada [Phased Implementation Plan](./PHASED-IMPLEMENTATION-PLAN.md), bukan pada visual mockup.

## 10. Prinsip Pengembangan

- Satu sumber data operasional; dashboard tidak memiliki data bisnis bayangan.
- Semua agregasi dan drill-down menjaga tenant dan scope access.
- Setiap KPI mempunyai definisi, unit, period, owner, formula, dan lineage.
- Alert harus actionable: severity, aging, PIC, due date, recommendation, dan status tindak lanjut.
- Dashboard eksekutif harus menyediakan jalur dari ringkasan ke bukti proyek.
- Integrasi mengikuti adapter/interface; tidak mengunci domain pada satu vendor BI atau sistem eksternal.
