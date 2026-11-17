DROP INDEX IF EXISTS idx_contracts_ppk_user;
ALTER TABLE contracts
    DROP COLUMN IF EXISTS spse_tender_code,
    DROP COLUMN IF EXISTS sirup_code,
    DROP COLUMN IF EXISTS agreement_form,
    DROP COLUMN IF EXISTS ppk_user_id;
