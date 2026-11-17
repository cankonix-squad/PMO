package project

import (
	"errors"

	"github.com/harmanto-49/cankora/internal/shared/constants"
)

// ErrProjectNotFound is returned when a project lookup yields no result.
var ErrProjectNotFound = errors.New("project not found")

// ErrCodeTaken is returned when the project code is already used.
var ErrCodeTaken = errors.New("project code already in use")

// ErrInvalidProjectMasterData is returned when a project references master data
// outside the caller organization, soft-deleted rows, or unknown IDs.
var ErrInvalidProjectMasterData = errors.New("invalid project master data")

// ErrTaskNotFound is returned when a task lookup yields no result.
var ErrTaskNotFound = errors.New("task not found")

// ErrMilestoneNotFound is returned when a milestone lookup yields no result.
var ErrMilestoneNotFound = errors.New("milestone not found")

// ErrIssueNotFound is returned when an issue lookup yields no result.
var ErrIssueNotFound = errors.New("issue not found")

// ErrRiskNotFound is returned when a risk lookup yields no result.
var ErrRiskNotFound = errors.New("risk not found")

// ErrBudgetNotFound is returned when a budget line lookup yields no result.
var ErrBudgetNotFound = errors.New("budget not found")

// ErrVendorNotFound is returned when a vendor lookup yields no result.
var ErrVendorNotFound = errors.New("vendor not found")

// ErrContractNotFound is returned when a contract lookup yields no result.
var ErrContractNotFound = errors.New("contract not found")

// ErrContractNumberTaken is returned when the contract number already exists
// within the same organization.
var ErrContractNumberTaken = errors.New("contract number already in use")

// ErrInvalidVendorType is returned when a vendor type is not VENDOR/CONSULTANT.
var ErrInvalidVendorType = errors.New("vendor type must be VENDOR or CONSULTANT")

// ErrInvalidContractStatus is returned when a contract status is not one of
// DRAFT/ACTIVE/AMENDED/COMPLETED/TERMINATED.
var ErrInvalidContractStatus = errors.New("invalid contract status")

// ErrInvalidContractDates is returned when start_date is after end_date.
var ErrInvalidContractDates = errors.New("contract start_date cannot be after end_date")

// ErrInvalidPPK is returned when ppk_user_id is not an active internal user of the organization.
var ErrInvalidPPK = errors.New("ppk_user_id must be an active internal user of this organization")

// ErrInvalidAgreementForm is returned for an unknown agreement_form.
var ErrInvalidAgreementForm = errors.New("agreement_form must be SPK, SURAT_PERJANJIAN, SURAT_PESANAN or BUKTI_PEMBELIAN")

// ErrInvalidContractValue is returned when contract_value is negative.
var ErrInvalidContractValue = errors.New("contract value cannot be negative")

// ErrVendorInUse is returned when deleting a vendor that is referenced by an
// active contract.
var ErrVendorInUse = errors.New("vendor is referenced by contracts and cannot be deleted")

// ErrDocumentNotFound is returned when a document lookup yields no result.
var ErrDocumentNotFound = errors.New("document not found")

// ErrDocumentMissingFile is returned when an upload request has no file part.
var ErrDocumentMissingFile = errors.New("no file uploaded")

// ErrDocumentStorage is returned when reading/writing the underlying file
// store fails (file missing on disk, permission error, etc.).
var ErrDocumentStorage = errors.New("document storage error")

// ErrCorrectiveActionNotFound is returned when a corrective action lookup yields no result.
var ErrCorrectiveActionNotFound = errors.New("corrective action not found")

var validVendorTypes = map[string]bool{
	constants.VendorTypeVendor:     true,
	constants.VendorTypeConsultant: true,
}

var validContractStatuses = map[string]bool{
	constants.ContractStatusDraft:      true,
	constants.ContractStatusActive:     true,
	constants.ContractStatusAmended:    true,
	constants.ContractStatusCompleted:  true,
	constants.ContractStatusTerminated: true,
}
