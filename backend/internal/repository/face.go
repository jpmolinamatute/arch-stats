package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jpmolinamatute/arch-stats/backend/internal/model"
)

var faceColumns = []string{
	"face_type",
	"face_name",
	"viewBox",
	"render_cross",
}

// FaceRepo manages target face definitions and scoring zones.
type FaceRepo struct {
	db      DBTX
	catalog []model.FaceRead
}

// NewFaceRepo constructs a FaceRepo. Accepts an optional DBTX to maintain constructor parity
// across repository instances.
func NewFaceRepo(db ...DBTX) *FaceRepo {
	var conn DBTX
	if len(db) > 0 {
		conn = db[0]
	}
	return &FaceRepo{
		db:      conn,
		catalog: DefaultFaceCatalog,
	}
}

// WithTx returns a new FaceRepo bound to the given transaction.
func (r *FaceRepo) WithTx(tx pgx.Tx) *FaceRepo {
	return &FaceRepo{
		db:      tx,
		catalog: r.catalog,
	}
}

// BuildSelectQuery constructs a squirrel SELECT statement targeting target face columns.
// Used for query validation and future database-backed catalog operations.
func (r *FaceRepo) BuildSelectQuery(faceType *model.FaceType) (sql string, args []any, err error) {
	q := StmtBuilder.Select(faceColumns...).From("face")
	if faceType != nil {
		q = q.Where(squirrel.Eq{"face_type": *faceType})
	}
	sql, args, err = q.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("building select face query: %w", err)
	}
	return sql, args, nil
}

// FindAll returns all available target face definitions.
func (r *FaceRepo) FindAll(ctx context.Context) ([]model.FaceRead, error) {
	_ = ctx
	res := make([]model.FaceRead, len(r.catalog))
	copy(res, r.catalog)
	return res, nil
}

// FindByType returns target face definitions matching the given face type.
// Returns an empty slice if no matching definition is found.
func (r *FaceRepo) FindByType(ctx context.Context, faceType model.FaceType) ([]model.FaceRead, error) {
	_ = ctx
	res := make([]model.FaceRead, 0)
	for _, f := range r.catalog {
		if f.FaceType == faceType {
			res = append(res, f)
		}
	}
	return res, nil
}

// FindByID retrieves a target face definition by its string identifier (face_type).
// Returns nil, nil if no matching definition exists.
func (r *FaceRepo) FindByID(ctx context.Context, id string) (*model.FaceRead, error) {
	_ = ctx
	for _, f := range r.catalog {
		if string(f.FaceType) == id {
			found := f
			return &found, nil
		}
	}
	return nil, nil
}
