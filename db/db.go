package db

import (
	"errors"
	"log"
	"signachurn/scan/proto"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

type DB struct {
	pb         *pocketbase.PocketBase
	Signatures *models.Collection
	Repos      *models.Collection
	Tags       *models.Collection
}

func NewDB() *DB {
	result := new(DB)
	result.pb = pocketbase.New()
	return result
}

func (db *DB) Start(Init func()) error {
	db.pb.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		err := db.EnsureCollections()
		if err != nil {
			log.Println("Issue ensuring collections", err)
			return err
		}
		Init()
		return nil
	})

	go func() {
		err := db.pb.Start()
		if err != nil {
			panic(err)
			// return err
		}
	}()

	return nil
}

func (db *DB) EnsureCollection(c *models.Collection) (*models.Collection, error) {
	result, err := db.pb.Dao().FindCollectionByNameOrId(c.Name)
	if result != nil && err == nil {
		return result, err
	}
	err = db.pb.Dao().SaveCollection(c)
	if err != nil {
		return nil, err
	}
	return db.pb.Dao().FindCollectionByNameOrId(c.Name)
}

func (db *DB) EnsureCollections() error {

	repos, err := db.EnsureCollection(&models.Collection{
		Name:       "repos",
		Type:       models.CollectionTypeBase,
		ListRule:   types.Pointer(""),
		ViewRule:   types.Pointer(""),
		CreateRule: nil, //only admins
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "url",
				Type:     schema.FieldTypeUrl,
				Required: true,
			},
		),
	})
	if err != nil {
		return err
	}
	db.Repos = repos

	sigs, err := db.EnsureCollection(&models.Collection{
		Name:       "signatures",
		Type:       models.CollectionTypeBase,
		ListRule:   types.Pointer("@request.auth.id != ''"),
		ViewRule:   types.Pointer(""),
		CreateRule: nil, //only admins
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: true,
			},
			&schema.SchemaField{
				Name:     "typestring",
				Type:     schema.FieldTypeText,
				Required: true,
			},
		),
		Indexes: types.JsonArray[string]{
			"CREATE UNIQUE INDEX idx_signature_name ON signatures (name, typestring)",
		},
	})
	if err != nil {
		return err
	}
	db.Signatures = sigs

	tags, err := db.EnsureCollection(&models.Collection{
		Name:       "tags",
		Type:       models.CollectionTypeBase,
		ListRule:   types.Pointer(""),
		ViewRule:   types.Pointer(""),
		CreateRule: nil, //only admins
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: true,
			},
			&schema.SchemaField{
				Name:     "repo",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: schema.RelationOptions{
					MaxSelect:    types.Pointer(1),
					CollectionId: repos.Id,
				},
			},
			&schema.SchemaField{
				Name:     "signatures",
				Type:     schema.FieldTypeRelation,
				Required: false,
				Options: schema.RelationOptions{
					MaxSelect:    nil,
					CollectionId: sigs.Id,
				},
			},
		),
		Indexes: types.JsonArray[string]{
			"CREATE UNIQUE INDEX idx_tag ON signatures (name, repo)",
		},
	})
	if err != nil {
		return err
	}
	db.Tags = tags

	return nil
}

func (db *DB) AddTag(RepoId string, TagName string, SignatureIds ...string) error {
	rec := models.NewRecord(db.Tags)
	rec.Set("repo", RepoId)
	rec.Set("name", TagName)
	rec.Set("signatures", SignatureIds)
	return db.pb.Dao().SaveRecord(rec)
}

func (db *DB) AddSignatures(Signatures ...*proto.Signature) (error, []string) {
	errs := []error{}
	results := []string{}
	for _, Signature := range Signatures {
		rec := models.NewRecord(db.Signatures)
		rec.Set("name", Signature.Name)
		rec.Set("typestring", Signature.AsString)
		err := db.pb.Dao().SaveRecord(rec)
		results = append(results, rec.Id)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...), results
}

func (db *DB) Stop() error {
	return db.pb.DB().Close()
}
