package query

import (
	"log"
	"template/internal/model"

	"github.com/samber/do/v2"
	"github.com/samber/mo"
	"gorm.io/gorm"
)

type Queries struct {
	db *gorm.DB
}

func NewQueries(i do.Injector) *Queries {
	db := do.MustInvoke[*gorm.DB](i)

	_ = mo.TupleToResult(db, db.AutoMigrate(
		new(model.User),
		new(model.Permission),
		new(model.Role),
	)).MapErr(
		func(err error) (*gorm.DB, error) {
			log.Printf("[Fatal] AutoMigrate failed: %v\n", err)
			return nil, err
		},
	).MustGet()

	return &Queries{
		db: db,
	}
}
