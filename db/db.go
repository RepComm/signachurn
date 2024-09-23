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
	signatures *models.Collection
}

func NewDB() *DB {
	result := new(DB)
	result.pb = pocketbase.New()
	return result
}

func (db *DB) Start(Init func()) error {
	db.pb.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
		err := db.EnsureCollections()
		if err != nil {
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

func (db *DB) EnsureCollections() error {
	reposCollectionId := ""

	c, err := db.pb.Dao().FindCollectionByNameOrId("repos")
	if c == nil || err != nil {
		err = db.pb.Dao().SaveCollection(&models.Collection{
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
		c, err = db.pb.Dao().FindCollectionByNameOrId("repos")
	}
	reposCollectionId = c.Id

	var sigsCollectionId string

	c, err = db.pb.Dao().FindCollectionByNameOrId("signatures")
	if c == nil || err != nil {
		db.pb.Dao().SaveCollection(&models.Collection{
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
		c, err = db.pb.Dao().FindCollectionByNameOrId("signatures")
	}
	db.signatures = c
	log.Println("signatures collection", db.signatures)
	sigsCollectionId = c.Id

	c, err = db.pb.Dao().FindCollectionByNameOrId("tags")
	if c == nil || err != nil {
		db.pb.Dao().SaveCollection(&models.Collection{
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
						CollectionId: reposCollectionId,
					},
				},
				&schema.SchemaField{
					Name:     "signatures",
					Type:     schema.FieldTypeRelation,
					Required: false,
					Options: schema.RelationOptions{
						MaxSelect:    nil,
						CollectionId: sigsCollectionId,
					},
				},
			),
		})
	}
	return nil
}

func (db *DB) AddSignatures(TagName string, Signatures ...*proto.Signature) error {
	errs := []error{}
	for _, Signature := range Signatures {
		rec := models.NewRecord(db.signatures)
		log.Println(Signature.Name, Signature.AsString)
		rec.Set("name", Signature.Name)
		rec.Set("typestring", Signature.AsString)
		err := db.pb.Dao().SaveRecord(rec)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (db *DB) Stop() error {
	return db.pb.DB().Close()
}
