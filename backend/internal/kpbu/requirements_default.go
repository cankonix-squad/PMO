package kpbu

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Source verification flags for checklist rows.
const (
	SourceVerified   = "VERIFIED"   // read directly in the regulation text
	SourceSecondary  = "SECONDARY"  // secondary source (LKPP, associations, law firms)
	SourceUnverified = "UNVERIFIED" // common practice; must be checked against Peraturan LKPP 1/2025
)

// defaultRequirement is one row of the "Matriks Dokumen Badan Usaha dalam KPBU"
// agreed on 2026-11-17. Rows are seeded per organization and stay editable.
type defaultRequirement struct {
	Code, Category, Name, AppliesTo, BeforeStage, LegalBasis, Source string
	Mandatory                                                        bool
}

// DefaultRequirements is the seed checklist. BeforeStage is the stage a project
// cannot enter until the row has an approved document.
var DefaultRequirements = []defaultRequirement{
	// A. Pengajuan prakarsa (unsolicited only)
	{"A1", CatPrakarsa, "Surat minat (LOI) dan dokumen pendukung prakarsa", AppliesUnsolicited, StagePlanning, "Permen PPN 7/2023 jo. 9/2025", SourceSecondary, true},
	{"A2", CatPrakarsa, "Studi kelayakan (hukum-kelembagaan, teknis, ekonomi-komersial, lingkungan-sosial, risiko)", AppliesUnsolicited, StagePreparationFBC, "Perpres 38/2015 Ps.14 ayat 4; Permen PPN 7/2023", SourceVerified, true},
	{"A3", CatPrakarsa, "Bukti kemampuan keuangan pemrakarsa", AppliesUnsolicited, StagePlanning, "Perpres 38/2015 Ps.14 ayat 3", SourceVerified, true},
	{"A4", CatPrakarsa, "Bukti keterpaduan teknis dengan rencana induk sektor", AppliesUnsolicited, StagePlanning, "Perpres 38/2015 Ps.14 ayat 3", SourceVerified, true},
	{"A5", CatPrakarsa, "Pengalihan hak studi kelayakan dan HKI setelah kompensasi", AppliesUnsolicited, "", "Perpres 38/2015 Ps.14 ayat 7", SourceVerified, false},
	// B. Prakualifikasi
	{"B1", CatPrakualifikasi, "Legalitas badan usaha: akta pendirian, NIB, NPWP, pengesahan Kemenkumham", AppliesAll, StageContractSigned, "Peraturan LKPP 1/2025", SourceUnverified, true},
	{"B2", CatPrakualifikasi, "Perjanjian konsorsium dan penunjukan lead member (bila konsorsium)", AppliesAll, "", "Peraturan LKPP 1/2025", SourceUnverified, false},
	{"B3", CatPrakualifikasi, "Kemampuan keuangan: laporan keuangan teraudit, ekuitas, surat dukungan bank/lender", AppliesAll, StageContractSigned, "Peraturan LKPP 1/2025", SourceUnverified, true},
	{"B4", CatPrakualifikasi, "Pengalaman proyek infrastruktur sejenis", AppliesAll, StageContractSigned, "Peraturan LKPP 1/2025", SourceUnverified, true},
	{"B5", CatPrakualifikasi, "Pakta integritas dan pernyataan tidak dalam sanksi/daftar hitam", AppliesAll, StageContractSigned, "Peraturan LKPP 1/2025", SourceUnverified, true},
	{"B6", CatPrakualifikasi, "Dokumen hasil prakualifikasi (tahap wajib)", AppliesAll, StageContractSigned, "Perpres 38/2015 Ps.38 ayat 2", SourceVerified, true},
	// C. Penawaran
	{"C1", CatPenawaran, "Penawaran teknis: desain, metode pelaksanaan dan O&M, jadwal", AppliesAll, StageContractSigned, "Peraturan LKPP 1/2025", SourceUnverified, true},
	{"C2", CatPenawaran, "Penawaran finansial: tarif/AP, model keuangan, rencana pembiayaan", AppliesAll, StageContractSigned, "Peraturan LKPP 1/2025", SourceUnverified, true},
	{"C3", CatPenawaran, "Jaminan penawaran", AppliesAll, StageContractSigned, "Peraturan LKPP 1/2025 (nilai belum terverifikasi)", SourceSecondary, true},
	{"C4", CatPenawaran, "Dokumen Swiss challenge / right to match", AppliesUnsolicited, "", "Peraturan LKPP 1/2025; Perpres 38/2015 Ps.14 ayat 5", SourceSecondary, false},
	// D. Penetapan sampai penandatanganan
	{"D1", CatPenandatanganan, "Akta pendirian Badan Usaha Pelaksana (PT baru) oleh pemenang", AppliesAll, StageContractSigned, "Perpres 38/2015 Ps.1 angka 8", SourceVerified, true},
	{"D2", CatPenandatanganan, "Jaminan pelaksanaan (maksimal 5% nilai investasi)", AppliesAll, StageContractSigned, "Perpres 38/2015 (isi Perjanjian KPBU)", SourceVerified, true},
	{"D3", CatPenandatanganan, "Perjanjian KPBU yang ditandatangani PJPK dan BUP", AppliesAll, StageFinancialClose, "Perpres 38/2015 Ps.41", SourceVerified, true},
	{"D4", CatPenandatanganan, "Penjaminan pemerintah (PT PII), bila ada", AppliesAll, "", "Perpres 38/2015; Perpres 78/2010", SourceUnverified, false},
	// E. Pembiayaan dan pelaksanaan
	{"E1", CatPembiayaan, "Bukti financial close (paling lambat 12 bulan setelah penandatanganan)", AppliesAll, StageConstruction, "Perpres 38/2015 Ps.42", SourceVerified, true},
	{"E2", CatPembiayaan, "Laporan pelaksanaan konstruksi/O&M dan pemenuhan indikator layanan", AppliesAll, "", "Perpres 38/2015 Ps.13 ayat 4; Perjanjian KPBU", SourceSecondary, false},
	{"E3", CatPembiayaan, "Dokumen serah terima aset di akhir masa konsesi", AppliesAll, StageClosed, "Perjanjian KPBU (status aset)", SourceVerified, true},
}

// SeedDefaultRequirements inserts missing default rows for an organization.
// Existing rows (matched by code, including admin-edited ones) are left as is.
func SeedDefaultRequirements(db *gorm.DB, orgID uuid.UUID) (int, error) {
	created := 0
	for i, d := range DefaultRequirements {
		var existing DocumentRequirement
		err := db.Unscoped().Where("organization_id = ? AND code = ?", orgID, d.Code).First(&existing).Error
		if err == nil {
			continue // keep admin edits and deliberate deletions
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return created, err
		}
		var before *string
		if d.BeforeStage != "" {
			s := d.BeforeStage
			before = &s
		}
		r := DocumentRequirement{
			OrganizationID: orgID, Code: d.Code, Category: d.Category, Name: d.Name,
			AppliesTo: d.AppliesTo, RequiredBeforeStage: before, IsMandatory: d.Mandatory,
			LegalBasis: d.LegalBasis, SourceStatus: d.Source, SortOrder: (i + 1) * 10, IsActive: true,
		}
		if err := db.Create(&r).Error; err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}
