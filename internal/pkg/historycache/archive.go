package historycache

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Summary struct {
	Scope        string `json:"scope"`
	GameId       int64  `json:"gameId"`
	CreateTime   int64  `json:"createTime"`
	GameDuration int    `json:"gameDuration"`
	GameMode     string `json:"gameMode"`
	GameType     string `json:"gameType"`
	QueueId      int    `json:"queueId"`
	ChampionId   int    `json:"championId"`
	Kills        int    `json:"kills"`
	Deaths       int    `json:"deaths"`
	Assists      int    `json:"assists"`
	Win          bool   `json:"win"`
	HasDetail    bool   `json:"hasDetail"`
}
type Page struct {
	List     []Summary `json:"list"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
	Total    int       `json:"total"`
	HasNext  bool      `json:"hasNext"`
}
type Player struct {
	Scope      string `json:"scope"`
	Puuid      string `json:"puuid"`
	GameName   string `json:"gameName"`
	TagLine    string `json:"tagLine"`
	LastViewed int64  `json:"lastViewed"`
	Summaries  int    `json:"summaries"`
	Details    int    `json:"details"`
}
type Filter struct {
	Scope, Puuid   string
	Page, PageSize int
	Queue          *int
	Win            *bool
	From, To       int64
	DetailsOnly    bool
}

func (s *Store) Index(scope, puuid, name, tag string, list []Summary, generation uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if generation != s.generation {
		return ErrGeneration
	}
	if scope == "" || puuid == "" {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("INSERT INTO players VALUES(?,?,?,?,?) ON CONFLICT(scope,puuid) DO UPDATE SET name=CASE WHEN excluded.name='' THEN players.name ELSE excluded.name END,tag=CASE WHEN excluded.tag='' THEN players.tag ELSE excluded.tag END,viewed=excluded.viewed", scope, puuid, name, tag, time.Now().UnixMilli())
	if err != nil {
		tx.Rollback()
		_ = s.trimIndex()
		return err
	}
	for _, item := range list {
		item.Scope = scope
		item.HasDetail = false
		body, e := json.Marshal(item)
		if e != nil {
			return e
		}
		_, err = tx.Exec("INSERT INTO summaries VALUES(?,?,?,?,?,?,?) ON CONFLICT(scope,puuid,game_id) DO UPDATE SET created=excluded.created,queue=excluded.queue,win=excluded.win,body=excluded.body", scope, puuid, item.GameId, item.CreateTime, item.QueueId, item.Win, string(body))
		if err != nil {
			tx.Rollback()
			_ = s.trimIndex()
			return err
		}
	}
	return tx.Commit()
}
func (s *Store) Players(search string, page, size int) ([]Player, int, error) {
	return s.FindPlayers(search, "", "", page, size)
}

func (s *Store) FindPlayers(search, scope, puuid string, page, size int) ([]Player, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := []Player{}
	var total int
	pattern := "%" + strings.ToLower(search) + "%"
	where := "lower(p.name||'#'||p.tag) LIKE ?"
	args := []any{pattern}
	if scope != "" && puuid != "" {
		where += " AND p.scope=? AND p.puuid=?"
		args = append(args, scope, puuid)
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM players p WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(`SELECT p.scope,p.puuid,p.name,p.tag,p.viewed,
 (SELECT COUNT(*) FROM summaries s WHERE s.scope=p.scope AND s.puuid=p.puuid),
 (SELECT COUNT(*) FROM summaries s WHERE s.scope=p.scope AND s.puuid=p.puuid AND EXISTS(SELECT 1 FROM entries e WHERE e.scope=s.scope AND e.kind='detail' AND e.id=CAST(s.game_id AS TEXT)))
 FROM players p WHERE `+where+` ORDER BY p.viewed DESC,p.puuid,p.scope LIMIT ? OFFSET ?`, append(args, size, page*size)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var player Player
		if err = rows.Scan(&player.Scope, &player.Puuid, &player.GameName, &player.TagLine, &player.LastViewed, &player.Summaries, &player.Details); err != nil {
			return nil, 0, err
		}
		list = append(list, player)
	}
	return list, total, rows.Err()
}
func (s *Store) History(filter Filter) (Page, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := Page{List: []Summary{}, Page: filter.Page, PageSize: filter.PageSize}
	conditions := []string{"s.scope=?", "s.puuid=?"}
	args := []any{filter.Scope, filter.Puuid}
	if filter.Queue != nil {
		conditions = append(conditions, "s.queue=?")
		args = append(args, *filter.Queue)
	}
	if filter.Win != nil {
		conditions = append(conditions, "s.win=?")
		args = append(args, *filter.Win)
	}
	if filter.From > 0 {
		conditions = append(conditions, "s.created>=?")
		args = append(args, filter.From)
	}
	if filter.To > 0 {
		conditions = append(conditions, "s.created<=?")
		args = append(args, filter.To)
	}
	detailSQL := "EXISTS(SELECT 1 FROM entries e WHERE e.scope=s.scope AND e.kind='detail' AND e.id=CAST(s.game_id AS TEXT))"
	if filter.DetailsOnly {
		conditions = append(conditions, detailSQL)
	}
	where := strings.Join(conditions, " AND ")
	if err := s.db.QueryRow("SELECT COUNT(*) FROM summaries s WHERE "+where, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	query := "SELECT s.body," + detailSQL + " FROM summaries s WHERE " + where + " ORDER BY s.created DESC,s.game_id DESC LIMIT ? OFFSET ?"
	args = append(args, filter.PageSize, filter.Page*filter.PageSize)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var body string
		var has bool
		if err = rows.Scan(&body, &has); err != nil {
			return result, err
		}
		var item Summary
		if err = json.Unmarshal([]byte(body), &item); err != nil {
			return result, ErrCorrupt
		}
		item.HasDetail = has
		result.List = append(result.List, item)
	}
	result.HasNext = (filter.Page+1)*filter.PageSize < result.Total
	return result, rows.Err()
}
func (s *Store) MarkAvailability(page *Page) {
	for i := range page.List {
		item := &page.List[i]
		item.HasDetail = s.Has(item.Scope, "detail", strconv.FormatInt(item.GameId, 10))
	}
}
func PageID(puuid string, page, size int) string { return fmt.Sprintf("%s:%d:%d", puuid, page, size) }
