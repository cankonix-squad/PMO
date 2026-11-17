package kpbu

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// financeCRUD wires list/create/update/delete for one K2 child resource.
type financeCRUD[T financeRecord, R any] struct {
	h        *Handler
	entity   string // audit entity type, e.g. kpbu_financing
	param    string // child id route param
	list     func(orgID, projectID uuid.UUID) ([]T, error)
	save     func(orgID, projectID uuid.UUID, id *uuid.UUID, r R) (*T, error)
	del      func(orgID, projectID, id uuid.UUID) (*T, error)
	idOf     func(*T) uuid.UUID
	labelOf  func(*T) string
	notFound string
}

func (f financeCRUD[T, R]) ids(c *gin.Context, withChild bool) (uuid.UUID, *uuid.UUID, bool) {
	pid, ok := parseID(c, "id")
	if !ok || !withChild {
		return pid, nil, ok
	}
	cid, ok := parseID(c, f.param)
	return pid, &cid, ok
}

func (f financeCRUD[T, R]) List(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, _, ok := f.ids(c, false)
	if !ok {
		return
	}
	list, err := f.list(claims.OrganizationID, pid)
	if err != nil {
		f.h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, f.entity+" retrieved", list)
}

func (f financeCRUD[T, R]) write(c *gin.Context, withChild bool) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, cid, ok := f.ids(c, withChild)
	if !ok {
		return
	}
	var req R
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	v, err := f.save(claims.OrganizationID, pid, cid, req)
	if err != nil {
		f.h.writeErr(c, err, f.notFound)
		return
	}
	if withChild {
		f.h.record(c, claims, f.entity+".updated", f.entity, f.idOf(v), f.labelOf(v), v)
		response.OK(c, f.entity+" updated", v)
		return
	}
	f.h.record(c, claims, f.entity+".created", f.entity, f.idOf(v), f.labelOf(v), v)
	response.Created(c, f.entity+" created", v)
}

func (f financeCRUD[T, R]) Create(c *gin.Context) { f.write(c, false) }
func (f financeCRUD[T, R]) Update(c *gin.Context) { f.write(c, true) }

func (f financeCRUD[T, R]) Delete(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	pid, cid, ok := f.ids(c, true)
	if !ok {
		return
	}
	v, err := f.del(claims.OrganizationID, pid, *cid)
	if err != nil {
		f.h.writeErr(c, err, f.notFound)
		return
	}
	f.h.record(c, claims, f.entity+".deleted", f.entity, f.idOf(v), f.labelOf(v), nil)
	response.NoContent(c)
}

// FinanceRoutes exposes the K2 handlers for router registration.
type FinanceRoutes struct {
	Financing financeCRUD[Financing, FinancingRequest]
	Support   financeCRUD[GovernmentSupport, GovernmentSupportRequest]
	Payment   financeCRUD[AvailabilityPayment, AvailabilityPaymentRequest]
}

func (h *Handler) Finance() FinanceRoutes {
	return FinanceRoutes{
		Financing: financeCRUD[Financing, FinancingRequest]{
			h: h, entity: "kpbu_financing", param: "financingID", notFound: "KPBU financing not found",
			list: h.svc.ListFinancings, save: h.svc.SaveFinancing, del: h.svc.DeleteFinancing,
			idOf: func(v *Financing) uuid.UUID { return v.ID }, labelOf: func(v *Financing) string { return v.Source + " " + v.Provider },
		},
		Support: financeCRUD[GovernmentSupport, GovernmentSupportRequest]{
			h: h, entity: "kpbu_government_support", param: "supportID", notFound: "KPBU government support not found",
			list: h.svc.ListSupports, save: h.svc.SaveSupport, del: h.svc.DeleteSupport,
			idOf: func(v *GovernmentSupport) uuid.UUID { return v.ID }, labelOf: func(v *GovernmentSupport) string { return v.Type + " " + v.DecisionNumber },
		},
		Payment: financeCRUD[AvailabilityPayment, AvailabilityPaymentRequest]{
			h: h, entity: "kpbu_availability_payment", param: "paymentID", notFound: "KPBU availability payment not found",
			list: h.svc.ListPayments, save: h.svc.SavePayment, del: h.svc.DeletePayment,
			idOf: func(v *AvailabilityPayment) uuid.UUID { return v.ID },
			labelOf: func(v *AvailabilityPayment) string {
				return fmt.Sprintf("%04d-%02d", v.PeriodYear, v.PeriodMonth)
			},
		},
	}
}

// FinanceSummary GET /kpbu/projects/:id/finance-summary
func (h *Handler) FinanceSummary(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	sum, err := h.svc.FinanceSummary(claims.OrganizationID, id)
	if err != nil {
		h.writeErr(c, err, "KPBU project not found")
		return
	}
	response.OK(c, "KPBU finance summary retrieved", sum)
}
