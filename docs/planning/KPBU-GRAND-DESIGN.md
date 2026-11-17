# Grand Design — Modul Monitoring KPBU
# PMO Control Tower — Kementerian Pekerjaan Umum

**Versi**: 0.1.0 (draft)
**Tanggal**: 2026-11-17
**Status**: Draft untuk diskusi. Fase K1 diimplementasikan 2026-11-17 (lihat `DEVELOPMENT-BACKLOG.md`, PMO-KPBU-K1); K2–K5 belum.

> **Catatan sumber.** Dokumen ini disusun dari pemahaman umum tentang skema KPBU di Indonesia
> (Perpres 38/2015 beserta aturan turunannya), bukan dari dokumen resmi Kementerian PU.
> Hal-hal yang ditandai **[VERIFIKASI]** harus dicek dengan regulasi dan prosedur terbaru
> unit pembiayaan infrastruktur Kementerian PU sebelum dijadikan requirement final.
> Nama unit, istilah, dan format laporan dapat berubah, terutama setelah PUPR dipecah menjadi
> Kementerian PU dan Kementerian PKP pada 2024.

---

## 1. Latar Belakang

PMO Control Tower saat ini dirancang untuk proyek yang dibiayai APBN. Strukturnya adalah
kontrak konstruksi, budget, progres fisik/keuangan, dan health score. Proyek KPBU
(Kerjasama Pemerintah dengan Badan Usaha) punya karakter yang berbeda:

| Aspek | Proyek APBN (sudah ada) | Proyek KPBU (belum ada) |
|---|---|---|
| Siklus | Perencanaan → konstruksi → serah terima | Perencanaan → penyiapan → transaksi → konstruksi → operasi (15–30+ tahun) → serah terima aset |
| Pihak utama | Satker/Balai + kontraktor + konsultan | PJPK + Badan Usaha Pelaksana (BUP) + simpul KPBU + lender + penjamin |
| Uang pemerintah | Belanja modal per tahun anggaran | Dukungan (VGF), jaminan, atau Availability Payment selama masa konsesi |
| Ukuran sukses | Fisik/keuangan sesuai rencana | Tahapan tepat waktu, *financial close*, KPI layanan terpenuhi, kewajiban kontinjensi terkendali |
| Risiko | Register risiko proyek | Alokasi risiko kontraktual (pemerintah vs BUP) + kewajiban kontinjensi |

Tanpa modul khusus, data KPBU tersebar di dokumen, spreadsheet, dan laporan terpisah.
Pimpinan sulit melihat portofolio KPBU secara utuh: mana yang macet di penyiapan, mana yang
gagal mencapai *financial close*, berapa total kewajiban Availability Payment per tahun, dan
proyek operasi mana yang KPI-nya tidak terpenuhi.

## 2. Tujuan

1. **Satu sumber data portofolio KPBU** di lingkungan Kementerian PU, terhubung dengan
   PMO Control Tower yang sudah ada.
2. **Pantau tahapan KPBU** dari identifikasi sampai serah terima aset, dengan deteksi dini
   keterlambatan per tahap.
3. **Pantau kewajiban fiskal**: jadwal dan realisasi Availability Payment (AP), dukungan VGF,
   dan eksposur jaminan.
4. **Pantau kinerja layanan** pada masa operasi: KPI/SLA, pengurangan pembayaran (deduction),
   dan sanksi.
5. **Dukung keputusan pimpinan**: proyek mana yang perlu intervensi, keputusan apa yang
   menunggu, dan risiko pemerintah mana yang sedang terekspos.

**Di luar cakupan versi pertama:** model keuangan lengkap (perhitungan NPV/IRR, *Value for
Money*), proses pengadaan badan usaha secara elektronik, dan integrasi langsung ke sistem
Kemenkeu, Bappenas, atau PT PII. Semuanya bisa menyusul.

## 3. Pengguna dan Peran

| Peran | Kebutuhan utama | Peta ke role PMO |
|---|---|---|
| Pimpinan (Menteri/Dirjen/Direktur) | Ringkasan portofolio, proyek bermasalah, keputusan yang menunggu | `EXECUTIVE_VIEWER` |
| Unit pembiayaan infrastruktur / simpul KPBU **[VERIFIKASI nama unit]** | Kelola portofolio, validasi data, koordinasi lintas tahap | `PMO` |
| Tim PJPK / Direktorat teknis (mis. Ditjen SDA, Bina Marga, Cipta Karya) | Input tahapan, dokumen, monitoring kontrak | `PROJECT_MANAGER` |
| Balai/BBWS/BWS (lokasi proyek) | Input data lapangan, verifikasi konstruksi dan operasi | `PROJECT_OFFICER` |
| Auditor / Inspektorat | Jejak keputusan, dokumen, perubahan data | `AUDITOR` |
| Badan Usaha Pelaksana (opsional, fase lanjut) | Lapor KPI layanan, ajukan invoice AP | Role baru `BUP_REPORTER`, akses hanya ke proyeknya |

## 4. Siklus Hidup KPBU (FSM)

Tahapan dimodelkan sebagai *state machine*, sama seperti FSM proyek, task, dan risk yang
sudah ada (`workflow.CanTransition()`).

```
IDENTIFIED ──► PLANNING ──► PREPARATION_OBC ──► PREPARATION_FBC ──► TRANSACTION
                                                                        │
     ┌──────────────────────────────────────────────────────────────────┘
     ▼
CONTRACT_SIGNED ──► FINANCIAL_CLOSE ──► CONSTRUCTION ──► OPERATION ──► HANDOVER ──► CLOSED

Dari tahap mana pun sebelum CONTRACT_SIGNED: ──► ON_HOLD ──► (kembali ke tahap asal)
                                              ──► CANCELLED
Dari CONTRACT_SIGNED sampai OPERATION:        ──► TERMINATED (pengakhiran dini)
```

| Tahap | Keluaran utama | Gerbang (syarat lanjut) |
|---|---|---|
| `IDENTIFIED` | Usulan proyek, studi pendahuluan | Masuk daftar rencana KPBU **[VERIFIKASI]** |
| `PLANNING` | Konsultasi publik awal, penetapan PJPK | Penetapan sebagai proyek KPBU |
| `PREPARATION_OBC` | *Outline Business Case* | OBC disetujui |
| `PREPARATION_FBC` | *Final Business Case*, *market sounding*, usulan dukungan/jaminan | FBC disetujui, dukungan/jaminan mendapat persetujuan prinsip |
| `TRANSACTION` | Prakualifikasi, RFP, evaluasi, penetapan pemenang | Pemenang ditetapkan |
| `CONTRACT_SIGNED` | Perjanjian KPBU ditandatangani, BUP terbentuk | Perjanjian ditandatangani |
| `FINANCIAL_CLOSE` | Perjanjian pinjaman, ekuitas tersetor | *Financial close* tercapai dalam batas waktu perjanjian |
| `CONSTRUCTION` | Progres konstruksi | Sertifikat laik operasi / COD (*Commercial Operation Date*) |
| `OPERATION` | Layanan berjalan, KPI dipantau, pembayaran berjalan | Akhir masa konsesi |
| `HANDOVER` | Serah terima aset ke pemerintah | Berita acara serah terima |

Setiap transisi tercatat di audit log (`kpbu_project.transitioned`) beserta tanggal, aktor,
dan dokumen pendukung.

**Deteksi dini:** setiap tahap punya target tanggal. Kalau lewat, proyek ditandai terlambat dan
masuk Command Center, sama seperti milestone terlambat sekarang.

## 5. Model Data

Mengikuti konvensi PMO: PK `UUID`, `organization_id` di setiap tabel bisnis, soft delete
`deleted_at`, uang `NUMERIC(20,2)` + `currency`, timestamp `TIMESTAMPTZ`.

### 5.1 Entitas inti

**`kpbu_projects`**: satu baris per proyek KPBU.

| Kolom | Keterangan |
|---|---|
| `project_id` | FK opsional ke `projects`. Saat konstruksi, progres fisik bisa memakai modul proyek yang sudah ada |
| `code`, `name`, `sector` | Sektor: SDA (bendungan, irigasi, air baku), jalan, air minum, dll. |
| `stage` | Tahap FSM pada bagian 4 |
| `pjpk_name`, `pjpk_unit` | Penanggung Jawab Proyek Kerjasama |
| `initiative` | `SOLICITED` (prakarsa pemerintah) atau `UNSOLICITED` (prakarsa badan usaha) |
| `return_scheme` | `USER_CHARGE`, `AVAILABILITY_PAYMENT`, atau `HYBRID` |
| `investment_value` | Nilai investasi (capex) |
| `concession_years`, `concession_start`, `concession_end` | Masa kerja sama |
| `cod_target`, `cod_actual` | Tanggal mulai operasi komersial |
| `financial_close_target`, `financial_close_actual` | |
| `org_unit_id`, `region_id`, `river_basin_id`, `latitude`, `longitude` | Lokasi, memakai master data yang sudah ada |

**`kpbu_stage_milestones`**: target dan realisasi per tahap (tanggal rencana, tanggal aktual,
status, catatan, dokumen). Dasar perhitungan keterlambatan.

**`kpbu_parties`**: pihak-pihak dalam proyek.

| Kolom | Keterangan |
|---|---|
| `role` | `PJPK`, `BUP`, `SPONSOR`, `LENDER`, `GUARANTOR`, `CONSULTANT`, `INDEPENDENT_ENGINEER` |
| `vendor_id` | FK opsional ke `vendors` yang sudah ada |
| `name`, `share_pct` | Untuk BUP/sponsor: komposisi kepemilikan |

**`kpbu_government_support`**: dukungan dan jaminan pemerintah.

| Kolom | Keterangan |
|---|---|
| `type` | `VGF`, `PDF` (fasilitas penyiapan), `GUARANTEE`, `LAND_ACQUISITION`, `TAX_INCENTIVE`, `OTHER` |
| `provider` | Mis. Kemenkeu, PT PII **[VERIFIKASI]** |
| `amount`, `status`, `approved_at` | |

### 5.2 Keuangan dan pembayaran

**`kpbu_payment_schedules`**: jadwal Availability Payment per periode (tahun/semester/bulan)
selama masa konsesi. Kolom: periode, jumlah rencana, dasar perhitungan.

**`kpbu_payments`**: realisasi pembayaran per periode.

| Kolom | Keterangan |
|---|---|
| `schedule_id` | FK ke jadwal |
| `gross_amount` | Nilai sebelum pengurangan |
| `deduction_amount` | Pengurangan karena KPI tidak terpenuhi (dihitung dari bagian 5.3) |
| `net_amount` | `gross − deduction`, **dihitung backend**, tidak diinput manual |
| `status` | `DRAFT → SUBMITTED → VERIFIED → APPROVED → PAID` (+ `REJECTED`) |
| `invoice_no`, `paid_at`, dokumen | |

Pola ini sama dengan budget sekarang: nilai turunan (variance, status) selalu dihitung backend.

### 5.3 Kinerja layanan (masa operasi)

**`kpbu_kpis`**: definisi KPI dari perjanjian. Contoh untuk SDA: ketersediaan air baku
(m³/detik), kualitas air, waktu respons perbaikan, tingkat kebocoran. Kolom: nama, satuan,
target, ambang, bobot pengurangan pembayaran, periode ukur.

**`kpbu_kpi_measurements`**: hasil pengukuran per periode. Kolom: nilai, sumber (laporan BUP /
verifikasi independen), status validasi (`SUBMITTED → VERIFIED / DISPUTED`), dokumen.
Hanya pengukuran `VERIFIED` yang dipakai untuk menghitung pengurangan pembayaran.

### 5.4 Risiko dan kewajiban kontinjensi

**`kpbu_risk_allocations`**: matriks alokasi risiko dari perjanjian.

| Kolom | Keterangan |
|---|---|
| `category` | Lahan, desain, konstruksi, operasi, permintaan, tarif, politik, kahar, finansial |
| `allocated_to` | `GOVERNMENT`, `BUP`, `SHARED` |
| `mitigation`, `contingent_exposure` | Perkiraan kewajiban pemerintah kalau risiko terjadi |
| `status` | Dipantau / terjadi / tertutup |

Risiko operasional harian tetap memakai modul `risks` yang sudah ada, ditautkan ke proyek.

### 5.5 Dokumen

Memakai modul dokumen yang sudah ada, dengan kategori tambahan: `KPBU_PRE_FS`, `KPBU_OBC`,
`KPBU_FBC`, `KPBU_RFP`, `KPBU_AGREEMENT`, `KPBU_AMENDMENT`, `KPBU_GUARANTEE`, `KPBU_KPI_REPORT`,
`KPBU_INVOICE`, `KPBU_HANDOVER`.

## 6. Fitur per Level Control Tower

### Level 1 — Pimpinan (`/kpbu` ringkasan)
- Jumlah proyek dan nilai investasi per tahap (funnel dari identifikasi sampai operasi).
- Proyek terlambat per tahap, dan proyek yang belum mencapai *financial close* menjelang batas waktu.
- **Kewajiban AP per tahun** (grafik 10–30 tahun ke depan) dan total realisasi tahun berjalan.
- Total dukungan dan jaminan pemerintah yang aktif.
- Proyek operasi dengan KPI tidak terpenuhi dan nilai pengurangan pembayaran.
- Peta sebaran proyek KPBU, memakai peta GIS yang sudah ada dengan filter "KPBU".

### Level 2 — Portofolio/sektor
- Perbandingan per sektor (SDA, jalan, air minum, dll.) dan per skema pengembalian.
- Durasi rata-rata per tahap, untuk menemukan tahap yang paling sering macet.
- Eksposur kewajiban kontinjensi per kategori risiko.

### Level 3 — Detail proyek
- Linimasa tahapan (rencana vs realisasi) dengan dokumen per tahap.
- Profil pihak: PJPK, BUP, sponsor, lender, penjamin.
- Jadwal vs realisasi pembayaran, termasuk riwayat pengurangan.
- Dasbor KPI layanan per periode.
- Matriks alokasi risiko.
- Saat konstruksi: progres fisik memakai modul proyek dan snapshot yang sudah ada.

## 7. Integrasi dengan Modul yang Sudah Ada

| Modul PMO | Pemakaian di KPBU |
|---|---|
| Projects, tasks, milestones | Fase konstruksi memakai proyek biasa yang ditautkan lewat `kpbu_projects.project_id` |
| Snapshot dan health score | Health proyek KPBU = gabungan dimensi baru (tahapan, pembayaran, KPI) dengan dimensi yang ada. Formula tetap dikonfigurasi dan disetujui bisnis |
| Risks, issues, corrective action | Risiko dan isu operasional |
| Vendors | Data BUP, sponsor, konsultan |
| Dokumen | Dokumen per tahap |
| Data governance | Validasi resmi data pembayaran dan KPI sebelum masuk laporan |
| Command Center | Eskalasi keterlambatan tahap, *financial close* kritis, KPI gagal |
| Notifikasi | Pengingat target tahap, jatuh tempo pembayaran, laporan KPI |
| Audit log | Semua transisi tahap, persetujuan pembayaran, perubahan KPI |
| Reporting / Power BI | Dataset baru `kpbu-portfolio`, `kpbu-payments`, `kpbu-kpi` |
| GIS | Lapisan atau filter proyek KPBU |

## 8. Desain Teknis

**Backend**: modul baru `backend/internal/kpbu/` dengan pola handler/service/repository seperti
modul lain.

```
/api/v1/kpbu/projects                         GET, POST
/api/v1/kpbu/projects/:id                     GET, PUT, DELETE
/api/v1/kpbu/projects/:id/transition          POST   (FSM tahapan)
/api/v1/kpbu/projects/:id/stage-milestones    GET, POST, PUT
/api/v1/kpbu/projects/:id/parties             GET, POST, PUT, DELETE
/api/v1/kpbu/projects/:id/support             GET, POST, PUT, DELETE
/api/v1/kpbu/projects/:id/payment-schedules   GET, POST, PUT
/api/v1/kpbu/projects/:id/payments            GET, POST, PATCH /:paymentID/status
/api/v1/kpbu/projects/:id/kpis                GET, POST, PUT
/api/v1/kpbu/projects/:id/kpi-measurements    GET, POST, PATCH /:measurementID/status
/api/v1/kpbu/projects/:id/risk-allocations    GET, POST, PUT, DELETE
/api/v1/kpbu/summary                          GET    (Level 1)
/api/v1/kpbu/portfolio                        GET    (Level 2)
```

**Permission**: resource baru `kpbu_project`, `kpbu_payment`, `kpbu_kpi`. Wajib ditambahkan ke
slice `resources` di `cmd/seed/main.go`. Kalau tidak, semua role mendapat 403 (kejadian yang
sama dengan `snapshot`/`baseline`). Persetujuan pembayaran memakai action `approve` yang sudah ada.

**Aturan bisnis yang ditegakkan backend**
- `net_amount`, keterlambatan tahap, dan total kewajiban selalu dihitung backend.
- Pembayaran tidak bisa `APPROVED` kalau pengukuran KPI periode itu belum `VERIFIED`.
- Transisi tahap mengikuti FSM. Tahap tertentu mensyaratkan dokumen tertentu
  (mis. `PREPARATION_FBC → TRANSACTION` butuh dokumen FBC yang disetujui).
- Semua query difilter `organization_id`. Role `BUP_REPORTER` hanya bisa melihat proyeknya sendiri.

**Frontend**
- `/kpbu`: ringkasan Level 1.
- `/kpbu/portfolio`: Level 2.
- `/kpbu/[id]`: detail proyek dengan tab Tahapan, Pihak, Pembayaran, KPI, Risiko, Dokumen.
- Grup sidebar baru "KPBU".

## 9. Tahapan Implementasi

| Fase | Isi | Hasil |
|---|---|---|
| **K1 — Fondasi** | Tabel `kpbu_projects`, `kpbu_stage_milestones`, `kpbu_parties`; FSM tahapan; CRUD; dokumen per tahap; daftar dan detail proyek | Portofolio KPBU tercatat dan tahapannya terpantau |
| **K2 — Keuangan** | Dukungan/jaminan, jadwal AP, realisasi pembayaran dengan alur persetujuan | Kewajiban fiskal terlihat |
| **K3 — Kinerja layanan** | Definisi KPI, pengukuran, verifikasi, pengurangan pembayaran otomatis | Masa operasi terpantau |
| **K4 — Control Tower** | Ringkasan Level 1/2, Command Center, notifikasi, dataset reporting, lapisan GIS | Pimpinan bisa memantau dan memutuskan |
| **K5 — Lanjutan (opsional)** | Portal BUP untuk lapor KPI dan invoice, matriks risiko dengan kewajiban kontinjensi, integrasi eksternal | Data langsung dari BUP |

Setiap fase menyertakan migration, seed permission, smoke test, dan pembaruan dokumen,
sesuai gerbang fase di `PHASED-IMPLEMENTATION-PLAN.md`.

## 10. Hal yang Perlu Dikonfirmasi

1. Unit mana di Kementerian PU yang menjadi pemilik modul ini (simpul KPBU / unit pembiayaan infrastruktur) **[VERIFIKASI]**.
2. Daftar tahapan resmi dan nama dokumen per tahap menurut regulasi terbaru **[VERIFIKASI]**.
3. Format laporan KPBU yang sudah dipakai sekarang (ke pimpinan, Bappenas, Kemenkeu).
4. Apakah BUP akan diberi akses langsung (portal BUP), atau data tetap diinput oleh tim PJPK.
5. Daftar proyek KPBU yang ada saat ini, untuk data awal dan uji coba.
6. Apakah proyek KPBU jalan tol (BPJT) termasuk cakupan, atau hanya sektor di luar tol.
7. Contoh KPI layanan dan rumus pengurangan pembayaran dari perjanjian yang sudah berjalan.

## 11. Risiko Implementasi

| Risiko | Mitigasi |
|---|---|
| Requirement dari pemahaman umum, bukan dokumen resmi | Validasi bagian 4 dan 5 dengan unit pemilik sebelum K1 dimulai |
| Data perjanjian bersifat rahasia (nilai, klausul) | Akses berbasis role dan proyek, audit setiap akses dokumen, enkripsi penyimpanan dokumen |
| Rumus pengurangan pembayaran berbeda per perjanjian | KPI dan bobot dikonfigurasi per proyek, bukan rumus global |
| Masa konsesi panjang (puluhan tahun) | Data periodik disimpan per periode, laporan memakai snapshot tervalidasi |
| Duplikasi dengan sistem lain di Kemenkeu/Bappenas | Posisikan PMO sebagai alat monitoring internal Kementerian PU; integrasi menyusul di K5 |
