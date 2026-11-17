-- Migration: 000043_contract_ppk_procurement (VENDOR-PORTAL-002)
-- Contract PPK (Pejabat Pembuat Komitmen), agreement form and procurement
-- references. Perpres 16/2018 jo. 12/2021: Ps.11 (PPK controls the contract),
-- Ps.28 (agreement forms).

ALTER TABLE contracts
    ADD COLUMN IF NOT EXISTS ppk_user_id      UUID REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS agreement_form   VARCHAR(30),
    ADD COLUMN IF NOT EXISTS sirup_code       VARCHAR(50),
    ADD COLUMN IF NOT EXISTS spse_tender_code VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_contracts_ppk_user
    ON contracts (ppk_user_id) WHERE deleted_at IS NULL AND ppk_user_id IS NOT NULL;
