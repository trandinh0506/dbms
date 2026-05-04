package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	mssql "github.com/microsoft/go-mssqldb"
)

type SessionManager struct {
	sessions map[string]*sql.Tx
	mu       sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*sql.Tx),
		mu:       sync.Mutex{},
	}
}

// 1. Start Session: Mở TX và trả về Session ID (UUID)
func (sm *SessionManager) StartSession(db *sql.DB) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	sessionID := uuid.New().String()
	sm.mu.Lock()
	sm.sessions[sessionID] = tx
	sm.mu.Unlock()

	go func() {
		time.Sleep(15 * time.Minute)
		sm.CloseSession(sessionID, "ROLLBACK")
	}()

	return sessionID, nil
}

// 2. Update Draft: Thực thi lệnh trên TX hiện có
func (sm *SessionManager) UpdateDraft(sessionID string, sid, cid int, mark float64) error {
	sm.mu.Lock()
	tx, exists := sm.sessions[sessionID]
	sm.mu.Unlock()

	if !exists {
		return errors.New("Session expired or not found")
	}

	query := `SET LOCK_TIMEOUT 50; 
	UPDATE ENROLLMENT WITH (ROWLOCK) SET mark = @p1 WHERE student_id = @p2 AND course_id = @p3`

	_, err := tx.Exec(query, mark, sid, cid)
	if err != nil {
		if isLockTimeoutError(err) {
			return errors.New("Hệ thống đang bận: Vui lòng đợi trong giây lát!")
		}
		if isDeadlockError(err) {
			return errors.New("Đã xảy ra deadlock: Vui lòng thử lại!")
		}
		return err
	}
	return nil
}

func (sm *SessionManager) UpdateDraftUnsafe(sessionID string, sid, cid int, mark float64) error {
	sm.mu.Lock()
	tx, exists := sm.sessions[sessionID]
	sm.mu.Unlock()

	if !exists {
		return errors.New("Session expired or not found")
	}

	query := ` 
	UPDATE ENROLLMENT WITH (ROWLOCK) SET mark = @p1 WHERE student_id = @p2 AND course_id = @p3`

	_, err := tx.Exec(query, mark, sid, cid)
	if err != nil {
		if isLockTimeoutError(err) {
			return errors.New("Hệ thống đang bận: Vui lòng đợi trong giây lát!")
		}
		if isDeadlockError(err) {
			return errors.New("Đã xảy ra deadlock: Vui lòng thử lại!")
		}
		return err
	}

	return nil
}

// 3. Close Session: Commit hoặc Rollback
func (sm *SessionManager) CloseSession(sessionID string, action string) error {
	sm.mu.Lock()
	tx, exists := sm.sessions[sessionID]
	if !exists {
		return errors.New("Session expired")
	}
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()

	if action == "COMMIT" {
		return tx.Commit()
	}
	return tx.Rollback()
}

func isMSSQLError(err error, number int32) bool {
	if err == nil {
		return false
	}
	var mssqlErr mssql.Error
	if errors.As(err, &mssqlErr) {
		return mssqlErr.Number == number
	}
	return strings.Contains(err.Error(), fmt.Sprintf("%d", number))
}

func isDeadlockError(err error) bool {
	return isMSSQLError(err, 1205)
}

func isLockTimeoutError(err error) bool {
	return isMSSQLError(err, 1222) || strings.Contains(err.Error(), "Lock request time out")
}
