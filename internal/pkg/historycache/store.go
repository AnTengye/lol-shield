// Package historycache 管理可丢弃、配额受限的离线战绩，不保存客户端认证信息。
package historycache

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	_ "modernc.org/sqlite"
)

const MaxBytes int64 = 500_000_000
const MaxJSONBytes = 32_000_000

var ErrMiss = errors.New("本地尚未保存该记录")
var ErrQuota = errors.New("本地缓存容量不足，本次未保存")
var ErrGeneration = errors.New("缓存已经清理，取消旧请求写入")
var ErrCorrupt = errors.New("本地记录损坏，请连接客户端重新获取")

type Store struct {
	mu               sync.Mutex
	db               *sql.DB
	lock             *flock.Flock
	dir              string
	limit, blobLimit int64
	generation       uint64
	closed           bool
}
type Record struct {
	Body        []byte
	ContentType string
	FetchedAt   time.Time
}
type Stats struct {
	UsedBytes int64  `json:"usedBytes"`
	MaxBytes  int64  `json:"maxBytes"`
	BlobBytes int64  `json:"blobBytes"`
	Details   int    `json:"details"`
	Summaries int    `json:"summaries"`
	Assets    int    `json:"assets"`
	Players   int    `json:"players"`
	Healthy   bool   `json:"healthy"`
	Message   string `json:"message,omitempty"`
}

// limit 可注入以验证淘汰；正式运行禁止超过产品上限。
func Open(dir string, limit int64) (*Store, error) {
	if limit <= 0 || limit > MaxBytes {
		limit = MaxBytes
	}
	for _, path := range []string{dir, filepath.Join(dir, "objects"), filepath.Join(dir, "index.sqlite"), filepath.Join(dir, "cache.lock")} {
		if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("缓存路径不能是符号链接")
		} else if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "objects"), 0700); err != nil {
		return nil, err
	}
	lock := flock.New(filepath.Join(dir, "cache.lock"))
	ok, err := lock.TryLock()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("缓存正被另一个应用进程使用")
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "index.sqlite"))
	if err != nil {
		lock.Close()
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, lock: lock, dir: dir, limit: limit, blobLimit: limit * 90 / 100}
	indexLimit := limit * 4 / 100
	if indexLimit < 65536 {
		s.Close()
		return nil, errors.New("测试配额至少需要 2 MB")
	}
	schema := fmt.Sprintf(`PRAGMA journal_mode=DELETE; PRAGMA synchronous=FULL; PRAGMA temp_store=MEMORY;
 PRAGMA auto_vacuum=INCREMENTAL; PRAGMA busy_timeout=3000; PRAGMA max_page_count=%d;
 CREATE TABLE IF NOT EXISTS entries(scope TEXT NOT NULL, kind TEXT NOT NULL, id TEXT NOT NULL, hash TEXT NOT NULL, content_type TEXT NOT NULL, fetched INTEGER NOT NULL, accessed INTEGER NOT NULL, PRIMARY KEY(scope,kind,id));
 CREATE INDEX IF NOT EXISTS entries_lru ON entries(kind,accessed);
 CREATE TABLE IF NOT EXISTS players(scope TEXT NOT NULL, puuid TEXT NOT NULL, name TEXT NOT NULL, tag TEXT NOT NULL, viewed INTEGER NOT NULL, PRIMARY KEY(scope,puuid));
 CREATE TABLE IF NOT EXISTS summaries(scope TEXT NOT NULL, puuid TEXT NOT NULL, game_id INTEGER NOT NULL, created INTEGER NOT NULL, queue INTEGER NOT NULL, win INTEGER NOT NULL, body TEXT NOT NULL, PRIMARY KEY(scope,puuid,game_id));
 CREATE INDEX IF NOT EXISTS summaries_player ON summaries(scope,puuid,created DESC,game_id DESC);
 PRAGMA user_version=1;`, indexLimit/4096)
	if _, err = db.Exec(schema); err != nil {
		s.Close()
		return nil, err
	}
	var pageSize, pageCount int64
	if err = db.QueryRow("PRAGMA page_size").Scan(&pageSize); err != nil {
		s.Close()
		return nil, err
	}
	if err = db.QueryRow("PRAGMA page_count").Scan(&pageCount); err != nil {
		s.Close()
		return nil, err
	}
	if pageSize != 4096 || pageSize*pageCount > indexLimit {
		s.Close()
		return nil, ErrQuota
	}
	var integrity string
	if err = db.QueryRow("PRAGMA quick_check").Scan(&integrity); err != nil || integrity != "ok" {
		s.Close()
		return nil, ErrCorrupt
	}
	if err = s.recover(); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	err := s.db.Close()
	_ = s.lock.Close()
	return err
}
func (s *Store) Generation() uint64 { s.mu.Lock(); defer s.mu.Unlock(); return s.generation }
func hashBody(body []byte) string   { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }
func (s *Store) objectPath(hash string) (string, error) {
	if len(hash) != 64 {
		return "", ErrCorrupt
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return "", ErrCorrupt
	}
	return filepath.Join(s.dir, "objects", hash), nil
}
func (s *Store) Read(scope, kind, id string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var hash, contentType string
	var fetched int64
	err := s.db.QueryRow("SELECT hash,content_type,fetched FROM entries WHERE scope=? AND kind=? AND id=?", scope, kind, id).Scan(&hash, &contentType, &fetched)
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, ErrMiss
	}
	if err != nil {
		return Record{}, err
	}
	path, err := s.objectPath(hash)
	if err != nil {
		return Record{}, err
	}
	body, err := readObject(path)
	if err != nil && !os.IsNotExist(err) && !errors.Is(err, ErrCorrupt) {
		return Record{}, err
	}
	if err != nil || hashBody(body) != hash {
		_, _ = s.db.Exec("DELETE FROM entries WHERE hash=?", hash)
		_ = s.gc()
		return Record{}, ErrCorrupt
	}
	if kind != "asset" {
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			_, _ = s.db.Exec("DELETE FROM entries WHERE hash=?", hash)
			_ = s.gc()
			return Record{}, ErrCorrupt
		}
		body, err = io.ReadAll(io.LimitReader(reader, MaxJSONBytes+1))
		_ = reader.Close()
		if err != nil || len(body) > MaxJSONBytes || !json.Valid(body) {
			_, _ = s.db.Exec("DELETE FROM entries WHERE hash=?", hash)
			_ = s.gc()
			return Record{}, ErrCorrupt
		}
	}
	_, _ = s.db.Exec("UPDATE entries SET accessed=? WHERE scope=? AND kind=? AND id=?", time.Now().UnixMilli(), scope, kind, id)
	return Record{body, contentType, time.UnixMilli(fetched)}, nil
}
func (s *Store) Put(scope, kind, id string, body []byte, contentType string, generation uint64) error {
	if scope == "" || id == "" {
		return errors.New("缺少缓存标识")
	}
	if len(body) > MaxJSONBytes || (kind == "asset" && len(body) > 2_000_000) {
		return ErrQuota
	}
	if kind != "asset" {
		if !json.Valid(body) {
			return errors.New("无效 JSON")
		}
		var out bytes.Buffer
		writer := gzip.NewWriter(&out)
		if _, err := writer.Write(body); err != nil {
			return err
		}
		if err := writer.Close(); err != nil {
			return err
		}
		body = out.Bytes()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if generation != s.generation {
		return ErrGeneration
	}
	if s.closed {
		return errors.New("缓存已关闭")
	}
	hash := hashBody(body)
	path, _ := s.objectPath(hash)
	info, statErr := os.Lstat(path)
	if statErr != nil && !os.IsNotExist(statErr) {
		return statErr
	}
	if statErr == nil {
		if !info.Mode().IsRegular() {
			return ErrCorrupt
		}
		existing, err := readObject(path)
		if err != nil {
			return err
		}
		if hashBody(existing) != hash {
			return ErrCorrupt
		}
	}
	if os.IsNotExist(statErr) {
		if err := s.makeRoom(int64(len(body)), scope, kind, id); err != nil {
			return err
		}
		temp, err := os.CreateTemp(filepath.Join(s.dir, "objects"), "pending-")
		if err != nil {
			return err
		}
		tmpPath := temp.Name()
		defer os.Remove(tmpPath)
		if _, err = temp.Write(body); err != nil {
			temp.Close()
			return err
		}
		if err = temp.Sync(); err != nil {
			temp.Close()
			return err
		}
		if err = temp.Close(); err != nil {
			return err
		}
		if err = os.Rename(tmpPath, path); err != nil {
			return err
		}
	}
	now := time.Now().UnixMilli()
	_, err := s.db.Exec("INSERT INTO entries VALUES(?,?,?,?,?,?,?) ON CONFLICT(scope,kind,id) DO UPDATE SET hash=excluded.hash,content_type=excluded.content_type,fetched=excluded.fetched,accessed=excluded.accessed", scope, kind, id, hash, contentType, now, now)
	if err != nil {
		_ = s.trimIndex()
		_ = s.gc()
		return err
	}
	return s.gc()
}
func (s *Store) PutJSON(scope, kind, id string, data any, generation uint64) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.Put(scope, kind, id, body, "application/json", generation)
}
func (s *Store) Has(scope, kind, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM entries WHERE scope=? AND kind=? AND id=?", scope, kind, id).Scan(&n)
	return n > 0
}
func (s *Store) makeRoom(incoming int64, protectedScope, protectedKind, protectedID string) error {
	if incoming > s.blobLimit {
		return ErrQuota
	}
	used, err := directoryBytes(filepath.Join(s.dir, "objects"))
	if err != nil {
		return err
	}
	if used+incoming <= s.blobLimit {
		return nil
	}
	target := max(int64(0), s.limit*80/100-incoming)
	for used > target {
		var scope, kind, id string
		err = s.db.QueryRow("SELECT scope,kind,id FROM entries WHERE NOT(scope=? AND kind=? AND id=?) ORDER BY CASE kind WHEN 'asset' THEN 0 WHEN 'detail' THEN 1 ELSE 2 END,accessed LIMIT 1", protectedScope, protectedKind, protectedID).Scan(&scope, &kind, &id)
		if err != nil {
			if used+incoming <= s.blobLimit {
				return nil
			}
			return ErrQuota
		}
		if _, err = s.db.Exec("DELETE FROM entries WHERE scope=? AND kind=? AND id=?", scope, kind, id); err != nil {
			return err
		}
		if err = s.gc(); err != nil {
			return err
		}
		used, err = directoryBytes(filepath.Join(s.dir, "objects"))
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) trimIndex() error {
	_, err := s.db.Exec(`DELETE FROM entries WHERE rowid IN (SELECT rowid FROM entries WHERE kind='page' ORDER BY accessed LIMIT 100);
 DELETE FROM summaries WHERE rowid IN (SELECT rowid FROM summaries ORDER BY created LIMIT 200);
 DELETE FROM players WHERE NOT EXISTS(SELECT 1 FROM summaries WHERE summaries.scope=players.scope AND summaries.puuid=players.puuid);
 PRAGMA incremental_vacuum(100);`)
	return err
}
func (s *Store) recover() error {
	rows, err := s.db.Query("SELECT DISTINCT hash FROM entries")
	if err != nil {
		return err
	}
	var missing []string
	for rows.Next() {
		var hash string
		if err = rows.Scan(&hash); err != nil {
			rows.Close()
			return err
		}
		path, e := s.objectPath(hash)
		if e != nil {
			missing = append(missing, hash)
		} else if info, statErr := os.Lstat(path); statErr != nil || !info.Mode().IsRegular() {
			missing = append(missing, hash)
		}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	for _, hash := range missing {
		if _, err = s.db.Exec("DELETE FROM entries WHERE hash=?", hash); err != nil {
			return err
		}
	}
	if err = s.gc(); err != nil {
		return err
	}
	used, err := directoryBytes(s.dir)
	if err != nil {
		return err
	}
	if used > s.limit {
		return ErrQuota
	}
	return nil
}

// gc 只删除对象目录内没有索引引用的文件，不追踪符号链接。
func (s *Store) gc() error {
	rows, err := s.db.Query("SELECT DISTINCT hash FROM entries")
	if err != nil {
		return err
	}
	refs := map[string]bool{}
	for rows.Next() {
		var hash string
		if err = rows.Scan(&hash); err != nil {
			rows.Close()
			return err
		}
		refs[hash] = true
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	files, err := os.ReadDir(filepath.Join(s.dir, "objects"))
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() && !refs[file.Name()] {
			if err = os.Remove(filepath.Join(s.dir, "objects", file.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
func readObject(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxJSONBytes+65536 {
		return nil, ErrCorrupt
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, MaxJSONBytes+65537))
}

func directoryBytes(dir string) (int64, error) {
	var size int64
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return nil
	})
	return size, err
}
func (s *Store) Stats() (Stats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := Stats{MaxBytes: s.limit, Healthy: true}
	var err error
	result.UsedBytes, err = directoryBytes(s.dir)
	if err != nil {
		return result, err
	}
	result.BlobBytes, err = directoryBytes(filepath.Join(s.dir, "objects"))
	if err != nil {
		return result, err
	}
	for _, q := range []struct {
		query string
		out   *int
	}{{"SELECT COUNT(*) FROM entries WHERE kind='detail'", &result.Details}, {"SELECT COUNT(*) FROM entries WHERE kind='asset'", &result.Assets}, {"SELECT COUNT(*) FROM summaries", &result.Summaries}, {"SELECT COUNT(*) FROM players", &result.Players}} {
		if err = s.db.QueryRow(q.query).Scan(q.out); err != nil {
			return result, err
		}
	}
	return result, nil
}
func (s *Store) Clear(kind string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if kind != "all" && kind != "assets" {
		return errors.New("无效清理范围")
	}
	s.generation++
	query := "DELETE FROM entries WHERE kind='asset'"
	if kind == "all" {
		query = "DELETE FROM entries; DELETE FROM summaries; DELETE FROM players;"
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(query); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if err := s.gc(); err != nil {
		return err
	}
	_, err = s.db.Exec("PRAGMA incremental_vacuum")
	return err
}

// SafeAssetPath 只允许图标资源，拒绝目录穿越、查询参数和外部地址。
func SafeAssetPath(path string) bool {
	if strings.Contains(path, "..") || strings.ContainsAny(path, "\\?#\x00") {
		return false
	}
	lower := strings.ToLower(path)
	for _, prefix := range []string{"/v1/champion-icons/", "/v1/profile-icons/", "/assets/items/icons2d/", "/data/spells/icons2d/", "/assets/ux/"} {
		if strings.HasPrefix(lower, prefix) {
			return strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".webp")
		}
	}
	return false
}
