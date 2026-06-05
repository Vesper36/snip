package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"github.com/vesper/snip/internal/models"
)

type Store struct{ db *sql.DB }

func New(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath+"?_journal=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(3)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Ping checks database connectivity.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) migrate() error {
	// Schema version tracking
	s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY)`)
	var version int
	s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)

	migrations := []struct {
		version int
		sql     string
	}{
		{
			1, `CREATE TABLE IF NOT EXISTS pastes (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			slug            TEXT NOT NULL UNIQUE,
			title           TEXT NOT NULL DEFAULT '',
			content         TEXT NOT NULL DEFAULT '',
			language        TEXT NOT NULL DEFAULT 'plaintext',
			file_size       INTEGER NOT NULL DEFAULT 0,
			is_encrypted    INTEGER NOT NULL DEFAULT 0,
			password_hash   TEXT NOT NULL DEFAULT '',
			burn_after_read INTEGER NOT NULL DEFAULT 0,
			expires_at      DATETIME,
			max_views       INTEGER NOT NULL DEFAULT 0,
			views           INTEGER NOT NULL DEFAULT 0,
			author_ip       TEXT NOT NULL DEFAULT '',
			created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		},
		{
			2, `CREATE TABLE IF NOT EXISTS api_tokens (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			name         TEXT NOT NULL,
			token_hash   TEXT NOT NULL UNIQUE,
			token_prefix TEXT NOT NULL,
			last_used_at DATETIME,
			expires_at   DATETIME,
			created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		},
		{3, `CREATE INDEX IF NOT EXISTS idx_pastes_slug ON pastes(slug)`},
		{4, `CREATE INDEX IF NOT EXISTS idx_pastes_expires ON pastes(expires_at)`},
		{5, `CREATE INDEX IF NOT EXISTS idx_pastes_created ON pastes(created_at DESC)`},
		{6, `CREATE INDEX IF NOT EXISTS idx_tokens_hash ON api_tokens(token_hash)`},
		{7, `CREATE INDEX IF NOT EXISTS idx_pastes_burn ON pastes(burn_after_read, views)`},
	}

	for _, m := range migrations {
		if m.version > version {
			if _, err := s.db.Exec(m.sql); err != nil {
				return fmt.Errorf("migration %d: %w", m.version, err)
			}
			s.db.Exec("INSERT INTO schema_version (version) VALUES (?)", m.version)
		}
	}
	return nil
}

// --- Paste Operations ---

func (s *Store) CreatePaste(p *models.Paste) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	res, err := s.db.Exec(`INSERT INTO pastes
		(slug,title,content,language,file_size,is_encrypted,password_hash,burn_after_read,expires_at,max_views,author_ip,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.Slug, p.Title, p.Content, p.Language, p.FileSize, p.IsEncrypted,
		p.PasswordHash, p.BurnAfterRead, p.ExpiresAt, p.MaxViews, p.AuthorIP, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) GetPaste(slug string) (*models.Paste, error) {
	p := &models.Paste{}
	err := s.db.QueryRow(`SELECT id,slug,title,content,language,file_size,is_encrypted,password_hash,
		burn_after_read,expires_at,max_views,views,author_ip,created_at,updated_at
		FROM pastes WHERE slug=?`, slug).Scan(
		&p.ID, &p.Slug, &p.Title, &p.Content, &p.Language, &p.FileSize, &p.IsEncrypted,
		&p.PasswordHash, &p.BurnAfterRead, &p.ExpiresAt, &p.MaxViews, &p.Views, &p.AuthorIP,
		&p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (s *Store) IncrementViews(slug string) error {
	_, err := s.db.Exec("UPDATE pastes SET views=views+1 WHERE slug=?", slug)
	return err
}

func (s *Store) DeletePaste(slug string) error {
	_, err := s.db.Exec("DELETE FROM pastes WHERE slug=?", slug)
	return err
}

func (s *Store) ListPastes(limit, offset int) ([]*models.Paste, error) {
	rows, err := s.db.Query(`SELECT id,slug,title,
		CASE WHEN password_hash != '' THEN '' ELSE content END as content,
		language,file_size,is_encrypted,password_hash,
		burn_after_read,expires_at,max_views,views,author_ip,created_at,updated_at
		FROM pastes WHERE (expires_at IS NULL OR expires_at>datetime('now'))
		ORDER BY created_at DESC LIMIT? OFFSET?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Paste
	for rows.Next() {
		p := &models.Paste{}
		rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Content, &p.Language, &p.FileSize, &p.IsEncrypted,
			&p.PasswordHash, &p.BurnAfterRead, &p.ExpiresAt, &p.MaxViews, &p.Views, &p.AuthorIP,
			&p.CreatedAt, &p.UpdatedAt)
		out = append(out, p)
	}
	return out, nil
}

func (s *Store) SearchPastes(q string, limit int) ([]*models.Paste, error) {
	// Escape LIKE wildcards to prevent injection
	escaped := strings.NewReplacer("%", "\\%", "_", "\\_", "\\", "\\\\").Replace(q)
	pat := "%" + escaped + "%"
	rows, err := s.db.Query(`SELECT id,slug,title,
		CASE WHEN password_hash != '' THEN '' ELSE content END as content,
		language,file_size,is_encrypted,password_hash,
		burn_after_read,expires_at,max_views,views,author_ip,created_at,updated_at
		FROM pastes WHERE (title LIKE? ESCAPE '\' OR content LIKE? ESCAPE '\') AND (expires_at IS NULL OR expires_at>datetime('now'))
		ORDER BY created_at DESC LIMIT?`, pat, pat, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Paste
	for rows.Next() {
		p := &models.Paste{}
		rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Content, &p.Language, &p.FileSize, &p.IsEncrypted,
			&p.PasswordHash, &p.BurnAfterRead, &p.ExpiresAt, &p.MaxViews, &p.Views, &p.AuthorIP,
			&p.CreatedAt, &p.UpdatedAt)
		out = append(out, p)
	}
	return out, nil
}

func (s *Store) CountPastes() (int, error) {
	var c int
	err := s.db.QueryRow("SELECT COUNT(*) FROM pastes WHERE expires_at IS NULL OR expires_at>datetime('now')").Scan(&c)
	return c, err
}

func (s *Store) CleanupExpired() (int64, error) {
	res, err := s.db.Exec(`DELETE FROM pastes WHERE
		(expires_at IS NOT NULL AND expires_at<datetime('now')) OR
		(burn_after_read=1 AND views>0)`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Backup creates a SQLite backup of the database.
func (s *Store) Backup(destPath string) error {
	// Use SQLite's VACUUM INTO for a consistent backup
	_, err := s.db.Exec("VACUUM INTO ?", destPath)
	return err
}

// --- API Token Operations ---

func (s *Store) CreateToken(t *models.APIToken) error {
	t.CreatedAt = time.Now()
	res, err := s.db.Exec(`INSERT INTO api_tokens(name,token_hash,token_prefix,expires_at,created_at) VALUES(?,?,?,?,?)`,
		t.Name, t.TokenHash, t.TokenPrefix, t.ExpiresAt, t.CreatedAt)
	if err != nil {
		return err
	}
	t.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) GetTokenByHash(hash string) (*models.APIToken, error) {
	t := &models.APIToken{}
	err := s.db.QueryRow(`SELECT id,name,token_hash,token_prefix,last_used_at,expires_at,created_at
		FROM api_tokens WHERE token_hash=?`, hash).Scan(
		&t.ID, &t.Name, &t.TokenHash, &t.TokenPrefix, &t.LastUsedAt, &t.ExpiresAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (s *Store) UpdateTokenUsed(id int64) error {
	_, err := s.db.Exec("UPDATE api_tokens SET last_used_at=? WHERE id=?", time.Now(), id)
	return err
}

func (s *Store) ListTokens() ([]*models.APIToken, error) {
	rows, err := s.db.Query(`SELECT id,name,token_hash,token_prefix,last_used_at,expires_at,created_at
		FROM api_tokens ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.APIToken
	for rows.Next() {
		t := &models.APIToken{}
		rows.Scan(&t.ID, &t.Name, &t.TokenHash, &t.TokenPrefix, &t.LastUsedAt, &t.ExpiresAt, &t.CreatedAt)
		out = append(out, t)
	}
	return out, nil
}

func (s *Store) DeleteToken(id int64) error {
	_, err := s.db.Exec("DELETE FROM api_tokens WHERE id=?", id)
	return err
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// --- Stats ---

type Stats struct {
	TotalPastes int   `json:"total_pastes"`
	TotalViews  int   `json:"total_views"`
	TotalTokens int   `json:"total_tokens"`
	TotalBytes  int64 `json:"total_bytes"`
}

func (s *Store) GetStats() (*Stats, error) {
	st := &Stats{}
	s.db.QueryRow(`SELECT COUNT(*),COALESCE(SUM(views),0),COALESCE(SUM(file_size),0) FROM pastes
		WHERE expires_at IS NULL OR expires_at>datetime('now')`).Scan(&st.TotalPastes, &st.TotalViews, &st.TotalBytes)
	s.db.QueryRow("SELECT COUNT(*) FROM api_tokens").Scan(&st.TotalTokens)
	return st, nil
}
