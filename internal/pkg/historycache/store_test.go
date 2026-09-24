package historycache

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir(), 2_000_000)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPersistenceIsolationAndLock(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir, 2_000_000)
	must(t, err)
	must(t, s.PutJSON("live:EUW", "detail", "42", map[string]int{"gameId": 42}, 0))
	must(t, s.Index("live:EUW", "p1", "玩家", "EU", []Summary{{GameId: 42, CreateTime: 100, Win: true}}, 0))
	if other, err := Open(dir, 2_000_000); err == nil {
		other.Close()
		t.Fatal("必须拒绝第二个写者")
	}
	must(t, s.Close())
	s, err = Open(dir, 2_000_000)
	must(t, err)
	defer s.Close()
	record, err := s.Read("live:EUW", "detail", "42")
	must(t, err)
	if string(record.Body) != `{"gameId":42}` {
		t.Fatal(string(record.Body))
	}
	for _, scope := range []string{"live:KR", "mock:default:EUW"} {
		if _, err := s.Read(scope, "detail", "42"); !errors.Is(err, ErrMiss) {
			t.Fatal(scope, err)
		}
	}
	players, total, err := s.Players("玩家#EU", 0, 20)
	must(t, err)
	if total != 1 || players[0].Details != 1 {
		t.Fatal(players, total)
	}
}
func TestArchiveDedupFiltersAndAvailability(t *testing.T) {
	s := openTestStore(t)
	items := []Summary{{GameId: 1, CreateTime: 200, QueueId: 450, Win: true}, {GameId: 2, CreateTime: 200, QueueId: 420, Win: false}, {GameId: 3, CreateTime: 100, QueueId: 450, Win: true}}
	must(t, s.Index("live:KR", "p", "玩家", "KR", items, 0))
	must(t, s.Index("live:KR", "p", "", "", items, 0))
	must(t, s.Index("live:KR", "other", "别人", "KR", items[:1], 0))
	must(t, s.PutJSON("live:KR", "detail", "1", map[string]int{"gameId": 1}, 0))
	page, err := s.History(Filter{Scope: "live:KR", Puuid: "p", PageSize: 2})
	must(t, err)
	if page.Total != 3 || !page.HasNext || page.List[0].GameId != 2 || !page.List[1].HasDetail {
		t.Fatal(page)
	}
	win, queue := true, 450
	page, err = s.History(Filter{Scope: "live:KR", Puuid: "p", PageSize: 20, Queue: &queue, Win: &win, From: 150, To: 250, DetailsOnly: true})
	must(t, err)
	if page.Total != 1 || page.List[0].GameId != 1 {
		t.Fatal(page)
	}
	_, err = s.db.Exec("DELETE FROM entries WHERE kind='detail'")
	must(t, err)
	page, err = s.History(Filter{Scope: "live:KR", Puuid: "p", PageSize: 20, DetailsOnly: true})
	must(t, err)
	if page.Total != 0 {
		t.Fatal(page)
	}
	players, total, err := s.Players("", 0, 20)
	must(t, err)
	if total != 2 || players[0].Details != 0 {
		t.Fatal(players)
	}
}
func TestQuotaEvictsImagesFirstAndAccountsJournal(t *testing.T) {
	s := openTestStore(t)
	must(t, s.PutJSON("live:KR", "detail", "1", map[string]int{"gameId": 1}, 0))
	for i := 0; i < 5; i++ {
		must(t, s.Put("assets", "asset", fmt.Sprint(i), bytes.Repeat([]byte{byte(i)}, 700_000), "image/png", 0))
		stats, err := s.Stats()
		must(t, err)
		if stats.UsedBytes > 2_000_000 || stats.BlobBytes > 1_800_000 {
			t.Fatal(stats)
		}
	}
	if !s.Has("live:KR", "detail", "1") || s.Has("assets", "asset", "0") {
		t.Fatal("LRU 图片必须先于详情淘汰")
	}
	tx, err := s.db.Begin()
	must(t, err)
	_, err = tx.Exec("UPDATE entries SET accessed=accessed+1")
	must(t, err)
	used, err := directoryBytes(s.dir)
	must(t, err)
	if used > s.limit {
		t.Fatalf("含事务日志占用超限 %d", used)
	}
	must(t, tx.Rollback())
	var journal string
	must(t, s.db.QueryRow("PRAGMA journal_mode").Scan(&journal))
	if journal != "delete" {
		t.Fatal(journal)
	}
	var pages int64
	must(t, s.db.QueryRow("PRAGMA max_page_count").Scan(&pages))
	if pages*4096 > 80_000 {
		t.Fatal(pages)
	}
}
func TestClearGenerationAndConcurrentAccess(t *testing.T) {
	s := openTestStore(t)
	must(t, s.PutJSON("live:KR", "detail", "1", map[string]int{"gameId": 1}, 0))
	must(t, s.Put("assets", "asset", "a", []byte("image"), "image/png", 0))
	must(t, s.Clear("assets"))
	if !s.Has("live:KR", "detail", "1") || s.Has("assets", "asset", "a") {
		t.Fatal("清理范围错误")
	}
	if err := s.PutJSON("live:KR", "detail", "2", map[string]int{}, 0); !errors.Is(err, ErrGeneration) {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			for j := 0; j < 10; j++ {
				generation := s.Generation()
				_ = s.PutJSON("live:KR", "detail", fmt.Sprint(i), map[string]int{"i": j}, generation)
				_, _ = s.Read("live:KR", "detail", fmt.Sprint(i))
				if i == 0 {
					_ = s.Clear("all")
				}
			}
		}(i)
	}
	group.Wait()
	generation := s.Generation()
	must(t, s.Clear("all"))
	if err := s.Index("live:KR", "p", "", "", nil, generation); !errors.Is(err, ErrGeneration) {
		t.Fatal(err)
	}
	stats, err := s.Stats()
	must(t, err)
	if stats.Details != 0 || stats.Players != 0 || stats.BlobBytes != 0 {
		t.Fatal(stats)
	}
}
func TestCorruptionRecoveryAndWriteFailurePreserveData(t *testing.T) {
	s := openTestStore(t)
	must(t, s.PutJSON("live:KR", "detail", "1", map[string]int{"gameId": 1}, 0))
	var hash string
	must(t, s.db.QueryRow("SELECT hash FROM entries").Scan(&hash))
	path, _ := s.objectPath(hash)
	must(t, os.WriteFile(path, []byte("损坏"), 0600))
	if _, err := s.Read("live:KR", "detail", "1"); !errors.Is(err, ErrCorrupt) {
		t.Fatal(err)
	}
	if s.Has("live:KR", "detail", "1") {
		t.Fatal("损坏引用应移除")
	}
	must(t, s.Put("assets", "asset", "old", bytes.Repeat([]byte{1}, 1_000_000), "image/png", 0))
	err := s.Put("assets", "asset", "old", bytes.Repeat([]byte{2}, 1_000_000), "image/png", 0)
	if !errors.Is(err, ErrQuota) {
		t.Fatal(err)
	}
	rec, err := s.Read("assets", "asset", "old")
	must(t, err)
	if rec.Body[0] != 1 {
		t.Fatal("失败不能覆盖旧数据")
	}
	must(t, os.WriteFile(filepath.Join(s.dir, "objects", "pending-abandoned"), []byte("临时文件"), 0600))
	must(t, s.recover())
	if _, err = os.Stat(filepath.Join(s.dir, "objects", "pending-abandoned")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err = s.Put("a", "asset", "huge", make([]byte, 2_000_001), "image/png", 0); !errors.Is(err, ErrQuota) {
		t.Fatal(err)
	}
	if err = s.Put("a", "detail", "bad", []byte("invalid"), "application/json", 0); err == nil {
		t.Fatal("无效JSON不能保存")
	}
}
func TestIndexPressureIsBounded(t *testing.T) {
	s := openTestStore(t)
	for i := 0; i < 500; i++ {
		_ = s.Index("live:KR", "p", "玩家", "KR", []Summary{{GameId: int64(i + 1), CreateTime: int64(i), GameMode: string(bytes.Repeat([]byte("x"), 500))}}, 0)
	}
	stats, err := s.Stats()
	must(t, err)
	if stats.UsedBytes > s.limit || stats.Summaries >= 500 {
		t.Fatal(stats)
	}
}
func TestReadOnlyAndCorruptDatabasePreserveRecords(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows ACL 在原生验收中覆盖")
	}
	s := openTestStore(t)
	must(t, s.PutJSON("live:KR", "detail", "1", map[string]int{"gameId": 1}, 0))
	var hash string
	must(t, s.db.QueryRow("SELECT hash FROM entries").Scan(&hash))
	path, _ := s.objectPath(hash)
	must(t, os.Chmod(path, 0000))
	_, err := s.Read("live:KR", "detail", "1")
	must(t, os.Chmod(path, 0600))
	if err == nil {
		t.Skip("当前用户不受文件权限限制")
	}
	if !s.Has("live:KR", "detail", "1") {
		t.Fatal("权限错误不能清除有效引用")
	}
	_, err = s.Read("live:KR", "detail", "1")
	must(t, err)
	dir := t.TempDir()
	bad := []byte("损坏的 SQLite 数据库不得静默删除")
	must(t, os.WriteFile(filepath.Join(dir, "index.sqlite"), bad, 0600))
	if cache, err := Open(dir, 2_000_000); err == nil {
		cache.Close()
		t.Fatal("应报告数据库损坏")
	}
	body, err := os.ReadFile(filepath.Join(dir, "index.sqlite"))
	must(t, err)
	if !bytes.Equal(body, bad) {
		t.Fatal("损坏数据库必须保留供恢复")
	}
}

func TestFullQuotaStress(t *testing.T) {
	if os.Getenv("SHIELD_CACHE_STRESS") != "1" {
		t.Skip("设置 SHIELD_CACHE_STRESS=1 执行真实 500 MB 压测")
	}
	s, err := Open(t.TempDir(), MaxBytes)
	must(t, err)
	defer s.Close()
	for i := 0; i < 300; i++ {
		body := bytes.Repeat([]byte{byte(i)}, 2_000_000)
		body[0], body[1] = byte(i), byte(i>>8)
		must(t, s.Put("resources", "asset", fmt.Sprint(i), body, "image/png", 0))
		stats, err := s.Stats()
		must(t, err)
		if stats.UsedBytes > MaxBytes || stats.BlobBytes > 450_000_000 {
			t.Fatal(stats)
		}
	}
	if s.Has("resources", "asset", "0") {
		t.Fatal("最旧图片必须被淘汰")
	}
}

func TestAssetPathAllowlist(t *testing.T) {
	for _, path := range []string{"/v1/champion-icons/1.png", "/ASSETS/Items/Icons2D/1001.png", "/DATA/Spells/Icons2D/flash.png"} {
		if !SafeAssetPath(path) {
			t.Fatal(path)
		}
	}
	for _, path := range []string{"/v1/champion-icons/../secret.png", "/v1/champion-icons/a.png?x", "/characters/skin.jpg", "https://evil/icon.png"} {
		if SafeAssetPath(path) {
			t.Fatal(path)
		}
	}
}
