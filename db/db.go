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
	psAddTag *sql.Stmt
}


type TagCommit struct {
	Short string
	Id string
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tag_info (
		id INTEGER PRIMARY KEY,
		short_name TEXT,
		commit_hash TEXT,
		repo_info_id INTEGER,
		FOREIGN KEY(repo_info_id) REFERENCES repo_info(id),
		UNIQUE(repo_info_id, commit_hash)
	)`)
	if err != nil {
		return nil, errors.Join(errors.New("failed to init db"), err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS commit_scan_job (
		id INTEGER PRIMARY KEY,
		tag_info_id INTEGER,
		repo_info_id INTEGER,
		FOREIGN KEY(repo_info_id) REFERENCES repo_info(id),
		FOREIGN KEY(tag_info_id) REFERENCES tag_info(id)
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

	result.psAddTag, err = db.Prepare(`INSERT INTO tag_info (short_name, commit_hash, repo_info_id) VALUES (?, ?, ?)`)
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

	err = db.psAddTag.Close()
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

func (db *DB) AddRepoByURL(url string) (repo_id int, err error) {
	res, err := db.psAddRepoByUrl.Exec(url)
	if err != nil {
		return 0, errors.Join(errors.New("failed to add repo by url"), err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Join(errors.New("failed to get id of added repo"), err)
	}
	return int(id), nil
}

func (db *DB) AddTag (tag TagCommit, repo_id int) error {
	_, err := db.psAddTag.Exec(tag.Short, tag.Id, repo_id)
	if err != nil {
		return errors.Join(errors.New("failed to add repo by url"), err)
	}
	return nil
}
