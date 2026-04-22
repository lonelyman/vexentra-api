// Package pgtx provides a small helper to thread a GORM transaction through
// context.Context, so that service-level "unit of work" spans multiple repositories
// without changing repo interfaces.
//
// Usage from service layer:
//
//	err := pgtx.Run(ctx, db, func(ctx context.Context) error {
//	    if err := projectRepo.Create(ctx, p); err != nil { return err }
//	    if err := memberRepo.Add(ctx, m);    err != nil { return err }
//	    return nil
//	})
//
// Inside a repository method:
//
//	tx := pgtx.DB(ctx, r.db)
//	if err := tx.WithContext(ctx).Create(m).Error; err != nil { ... }
package pgtx

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

// Run executes fn inside a GORM transaction. The callback receives a context
// carrying the transaction handle; repositories that call DB(ctx, fallback)
// will pick up the tx automatically.
//
// If fn returns an error, the transaction is rolled back.
func Run(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, ctxKey{}, tx))
	})
}

// DB returns the transaction stored in ctx if present, otherwise fallback.
// Repositories should funnel all queries through this so they transparently
// participate in a caller's transaction.
func DB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok {
		return tx
	}
	return fallback
}
