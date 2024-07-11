package db

import (
	"database/sql"
	"errors"

	_ "github.com/tursodatabase/go-libsql"
)

type DB struct {
	Addr string
	db *sql.DB

	psListRepos *sql.Stmt
	psAddRepoByUrl *sql.Stmt
}

func ConnectDB(addr string) (result *DB, err error) {
	result = new(DB)

	result.Addr = addr

	// connector,err := libsql.NewEmbeddedReplicaConnector(dbPath, primaryUrl)
	// if err != nil {
	// 	return nil, errors.Join(errors.New("failed to connect to db"),err)
	// }
	// result.db = sql.OpenDB(connector)

	db, err := sql.Open("libsql", addr)
	if err != nil {
		return nil, errors.Join(errors.New("failed to connect to db"), err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS repo_info (
		id INTEGER PRIMARY KEY,
		url TEXT,
		UNIQUE(url)
	)`)
	if err != nil {
		return nil, errors.Join(errors.New("failed to init db"), err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS commit_scan_job (
		id INTEGER PRIMARY KEY,
		commit_id TEXT,
		owning_repo_id INTEGER,
		FOREIGN KEY(owning_repo_id) REFERENCES repo_info(id)
	)`)
	if err != nil {
		return nil, errors.Join(errors.New("failed to init db"), err)
	}

	result.psListRepos, err = db.Prepare(`SELECT id, url FROM repo_info
	LIMIT ? OFFSET ?
	`)
	if err != nil {
		return nil, errors.Join(errors.New("failed to init db"), err)
	}

	result.psAddRepoByUrl, err = db.Prepare(`INSERT INTO repo_info (url) VALUES (?)`)
	if err != nil {
		return nil, errors.Join(errors.New("failed to init db"), err)
	}

	result.db = db

	return result, nil
}
func (db *DB) Close() error {
	_errs := make([]error, 0)
	err := db.psAddRepoByUrl.Close()
	if err != nil {
		_errs = append(_errs, err)
	}

	err = db.psListRepos.Close()
	if err != nil {
		_errs = append(_errs, err)
	}

	err = db.db.Close()
	if err != nil {
		_errs = append(_errs, err)
	}

	if len(_errs) > 0 {
		return errors.Join(_errs...)
	}
	return nil
}

type RepoInfo struct {
	Id int
	Url string
}

func (db *DB) ListRepos(count int, offset int) (results []RepoInfo, err error) {
	rows, err := db.psListRepos.Query(count, offset)
	if err != nil {
		return nil, errors.Join(errors.New("failed to fetch repos"), err)
	}
	results = make([]RepoInfo, 0)
	for rows.Next() {
		item := RepoInfo{}
		err = rows.Scan(&item.Id, &item.Url)
		if err != nil {
			return results, errors.Join(errors.New("error while fetching repo"), err)
		}
		results = append(results, item)
	}
	return results, nil
}

func (db *DB) AddRepoByURL(url string) error {
	_, err := db.psAddRepoByUrl.Exec(url)
	if err != nil {
		return errors.Join(errors.New("failed to add repo by url"), err)
	}
	return nil
}
