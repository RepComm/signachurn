package db

import (
	"errors"
	"log"
	"signachurn/scan/proto"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
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

func (db *DB) EnsureCollectionV(c *models.Collection) (*models.Collection, error) {
	result, err := db.pb.Dao().FindCollectionByNameOrId(c.Name)
	if result != nil && err == nil {
		return result, nil
	}
	f := forms.NewCollectionUpsert(db.pb, c)
	err = f.Submit()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (db *DB) EnsureCollections() error {

	repos, err := db.EnsureCollectionV(&models.Collection{
		Name:       "repos",
		Type:       models.CollectionTypeBase,
		ListRule:   types.Pointer(""),
		ViewRule:   types.Pointer(""),
		CreateRule: nil, //only admins
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:        "url",
				Type:        schema.FieldTypeUrl,
				Required:    true,
				Presentable: true,
			},
		),
	})
	if err != nil {
		return err
	}
	db.Repos = repos

	sigs, err := db.EnsureCollectionV(&models.Collection{
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

	tags, err := db.EnsureCollectionV(&models.Collection{
		Name:       "tags",
		Type:       models.CollectionTypeBase,
		ListRule:   types.Pointer(""),
		ViewRule:   types.Pointer(""),
		CreateRule: nil, //only admins
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(&schema.SchemaField{
			Name:        "name",
			Type:        schema.FieldTypeText,
			Required:    true,
			Presentable: true,
		}, &schema.SchemaField{
			Name:     "repo",
			Type:     schema.FieldTypeRelation,
			Required: true,
			Options: schema.RelationOptions{
				MaxSelect:    types.Pointer(1),
				CollectionId: repos.Id,
			},
		}, &schema.SchemaField{
			Name:     "signatures",
			Type:     schema.FieldTypeRelation,
			Required: false,
			Options: schema.RelationOptions{
				MaxSelect:    nil,
				CollectionId: sigs.Id,
			},
		}),
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

func (db *DB) EnsureRepo(RepoURL string) (string, error) {
	rec, err := db.pb.Dao().FindFirstRecordByData(
		db.Repos.Id,
		"url", RepoURL,
	)
	if rec != nil && err == nil {
		return rec.Id, nil
	}
	rec = models.NewRecord(db.Repos)
	rec.Set("url", RepoURL)
	f := forms.NewRecordUpsert(db.pb, rec)
	err = f.Submit()
	if err != nil {
		return "", err
	}
	return rec.Id, nil
}

func (db *DB) AddTag(RepoId string, TagName string, SignatureIds ...string) error {
	rec := models.NewRecord(db.Tags)
	rec.Set("repo", RepoId)
	rec.Set("name", TagName)
	rec.Set("signatures", SignatureIds)
	f := forms.NewRecordUpsert(db.pb, rec)
	return f.Submit()
}

func (db *DB) AddSignatures(Signatures ...*proto.Signature) (error, []string) {
	errs := []error{}
	results := []string{}
	for _, Signature := range Signatures {
		rec := models.NewRecord(db.Signatures)
		rec.Set("name", Signature.Name)
		rec.Set("typestring", Signature.AsString)
		f := forms.NewRecordUpsert(db.pb, rec)
		err := f.Submit()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		results = append(results, rec.Id)
	}
	return errors.Join(errs...), results
}

func (db *DB) Stop() error {
	return db.pb.DB().Close()
}
