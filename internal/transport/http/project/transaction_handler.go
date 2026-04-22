package projecthdl

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"vexentra-api/internal/modules/project"
	"vexentra-api/internal/modules/project/projectsvc"
	"vexentra-api/internal/modules/user"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	svc      projectsvc.TransactionService
	validate *validator.Validate
	logger   logger.Logger
}

func NewTransactionHandler(svc projectsvc.TransactionService, l logger.Logger) *TransactionHandler {
	if l == nil {
		l = logger.Get()
	}
	return &TransactionHandler{svc: svc, validate: validation.New(), logger: l}
}

// Create — POST /api/v1/projects/:id/transactions
func (h *TransactionHandler) Create(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	req := new(CreateTransactionRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "category_id ไม่ถูกต้อง")
	}

	t, svcErr := h.svc.Create(c.Context(), caller, projectID, projectsvc.CreateTransactionInput{
		CategoryID:   categoryID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		Note:         req.Note,
		OccurredAt:   req.OccurredAt,
	})
	if svcErr != nil {
		return svcErr
	}
	h.logger.Info("Transaction created", "projectID", projectID, "txID", t.ID)
	return presenter.RenderItem(c, NewTransactionResponse(t), fiber.StatusCreated)
}

// Get — GET /api/v1/projects/:id/transactions/:txID
func (h *TransactionHandler) Get(c fiber.Ctx) error {
	caller, projectID, txID, err := h.parseTxRoute(c)
	if err != nil {
		return err
	}
	t, svcErr := h.svc.Get(c.Context(), caller, projectID, txID)
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewTransactionResponse(t))
}

// List — GET /api/v1/projects/:id/transactions?category_id=&from=&to=&page=&limit=
func (h *TransactionHandler) List(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	q := presenter.ParseOffsetQuery(c)

	var categoryIDs []uuid.UUID
	if raw := strings.TrimSpace(c.Query("category_id", "")); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			id, perr := uuid.Parse(s)
			if perr != nil {
				return custom_errors.New(400, custom_errors.ErrValidation, "category_id ไม่ถูกต้อง")
			}
			categoryIDs = append(categoryIDs, id)
		}
	}

	var occurredGTE, occurredLT *time.Time
	if raw := strings.TrimSpace(c.Query("from", "")); raw != "" {
		t, terr := time.Parse(time.RFC3339, raw)
		if terr != nil {
			return custom_errors.New(400, custom_errors.ErrValidation, "from ต้องอยู่ในรูปแบบ RFC3339")
		}
		occurredGTE = &t
	}
	if raw := strings.TrimSpace(c.Query("to", "")); raw != "" {
		t, terr := time.Parse(time.RFC3339, raw)
		if terr != nil {
			return custom_errors.New(400, custom_errors.ErrValidation, "to ต้องอยู่ในรูปแบบ RFC3339")
		}
		occurredLT = &t
	}

	filter := project.TransactionFilter{
		CategoryIDs: categoryIDs,
		OccurredGTE: occurredGTE,
		OccurredLT:  occurredLT,
	}

	items, total, svcErr := h.svc.List(c.Context(), caller, projectID, filter, project.Pagination{
		Limit:  q.Limit,
		Offset: q.Offset,
	})
	if svcErr != nil {
		return svcErr
	}
	resp := make([]TransactionResponse, len(items))
	for i, t := range items {
		resp[i] = NewTransactionResponse(t)
	}
	pg := presenter.NewOffsetPagination(int(total), q.Limit, q.Offset)
	return presenter.RenderList(c, resp, pg)
}

// Summary — GET /api/v1/projects/:id/transactions/summary
func (h *TransactionHandler) Summary(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	totals, svcErr := h.svc.Summary(c.Context(), caller, projectID)
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewProjectTotalsResponse(totals))
}

// Update — PUT /api/v1/projects/:id/transactions/:txID
func (h *TransactionHandler) Update(c fiber.Ctx) error {
	caller, projectID, txID, err := h.parseTxRoute(c)
	if err != nil {
		return err
	}

	req := new(UpdateTransactionRequest)
	if err := c.Bind().Body(req); err != nil {
		return custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
	}
	if v := validation.Validate(h.validate, req); !v.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", v.Errors)
	}
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrValidation, "category_id ไม่ถูกต้อง")
	}

	t, svcErr := h.svc.Update(c.Context(), caller, projectID, txID, projectsvc.UpdateTransactionInput{
		CategoryID:   categoryID,
		Amount:       req.Amount,
		CurrencyCode: req.CurrencyCode,
		Note:         req.Note,
		OccurredAt:   req.OccurredAt,
	})
	if svcErr != nil {
		return svcErr
	}
	return presenter.RenderItem(c, NewTransactionResponse(t))
}

// Delete — DELETE /api/v1/projects/:id/transactions/:txID
func (h *TransactionHandler) Delete(c fiber.Ctx) error {
	caller, projectID, txID, err := h.parseTxRoute(c)
	if err != nil {
		return err
	}
	if svcErr := h.svc.Delete(c.Context(), caller, projectID, txID); svcErr != nil {
		return svcErr
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ExportCSV — GET /api/v1/projects/:id/transactions/export
// Streams all transactions for the project as a UTF-8 CSV (BOM-prefixed for Excel compat).
func (h *TransactionHandler) ExportCSV(c fiber.Ctx) error {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}

	rows, svcErr := h.svc.ListForExport(c.Context(), caller, projectID)
	if svcErr != nil {
		return svcErr
	}

	filename := fmt.Sprintf("transactions-%s.csv", projectID)
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.SendStreamWriter(func(w *bufio.Writer) {
		// BOM so Excel opens UTF-8 correctly without re-encoding
		w.WriteString("\xEF\xBB\xBF")
		w.WriteString("Occurred At,Category,Type,Amount,Currency,Note\n")
		for _, row := range rows {
			note := ""
			if row.Note != nil {
				note = strings.ReplaceAll(*row.Note, `"`, `""`)
				note = `"` + note + `"`
			}
			fmt.Fprintf(w, "%s,%s,%s,%s,%s,%s\n",
				row.OccurredAt.Format("2006-01-02"),
				csvEscape(row.CategoryName),
				row.CategoryType,
				row.Amount.StringFixed(2),
				row.CurrencyCode,
				note,
			)
		}
	})
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, `,"\n`) {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

// parseTxRoute parses caller + :id + :txID in one call; keeps per-handler noise low.
func (h *TransactionHandler) parseTxRoute(c fiber.Ctx) (user.Caller, uuid.UUID, uuid.UUID, error) {
	caller, err := auth.GetCaller(c)
	if err != nil {
		return caller, uuid.Nil, uuid.Nil, err
	}
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return caller, uuid.Nil, uuid.Nil, custom_errors.New(400, custom_errors.ErrInvalidFormat, "project id ไม่ถูกต้อง")
	}
	txID, err := uuid.Parse(c.Params("txID"))
	if err != nil {
		return caller, uuid.Nil, uuid.Nil, custom_errors.New(400, custom_errors.ErrInvalidFormat, "transaction id ไม่ถูกต้อง")
	}
	return caller, projectID, txID, nil
}
