package auth

import (
	"testing"
	"time"
)

func TestSessionManagerLifecycle(t *testing.T) {
	t.Parallel()

	sm := NewSessionManager()

	token, err := sm.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !sm.Validate(token) {
		t.Fatal("Validate() = false, want true")
	}

	sm.Delete(token)
	if sm.Validate(token) {
		t.Fatal("Validate() = true after Delete(), want false")
	}
}

func TestSessionManagerValidateRemovesExpiredSessions(t *testing.T) {
	t.Parallel()

	sm := NewSessionManager()
	sm.mu.Lock()
	sm.sessions["expired"] = time.Now().Add(-time.Minute)
	sm.mu.Unlock()

	if sm.Validate("expired") {
		t.Fatal("Validate() = true, want false for expired session")
	}

	sm.mu.RLock()
	_, exists := sm.sessions["expired"]
	sm.mu.RUnlock()
	if exists {
		t.Fatal("expired session still exists after Validate()")
	}
}

func TestSessionManagerResetInvalidatesAllSessions(t *testing.T) {
	t.Parallel()

	sm := NewSessionManager()
	token, err := sm.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	sm.Reset()

	if sm.Validate(token) {
		t.Fatal("Validate() = true after Reset(), want false")
	}
}
