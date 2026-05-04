package store

import "testing"

func TestInitAdminCreatesPasswordAndIsIdempotent(t *testing.T) {
	s := newTestStore(t)

	password, err := s.InitAdmin()
	if err != nil {
		t.Fatalf("InitAdmin() error = %v", err)
	}
	if len(password) != 16 {
		t.Fatalf("len(password) = %d, want 16", len(password))
	}

	valid, err := s.VerifyPassword(password)
	if err != nil {
		t.Fatalf("VerifyPassword(initial) error = %v", err)
	}
	if !valid {
		t.Fatal("VerifyPassword(initial) = false, want true")
	}

	secondPassword, err := s.InitAdmin()
	if err != nil {
		t.Fatalf("InitAdmin(second) error = %v", err)
	}
	if secondPassword != "" {
		t.Fatalf("InitAdmin(second) password = %q, want empty", secondPassword)
	}
}

func TestVerifyPasswordReturnsFalseWhenAdminMissing(t *testing.T) {
	s := newTestStore(t)

	valid, err := s.VerifyPassword("anything")
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
	if valid {
		t.Fatal("VerifyPassword() = true, want false")
	}
}

func TestUpdatePasswordReplacesExistingPassword(t *testing.T) {
	s := newTestStore(t)

	initialPassword, err := s.InitAdmin()
	if err != nil {
		t.Fatalf("InitAdmin() error = %v", err)
	}

	if err := s.UpdatePassword("new-password"); err != nil {
		t.Fatalf("UpdatePassword() error = %v", err)
	}

	oldValid, err := s.VerifyPassword(initialPassword)
	if err != nil {
		t.Fatalf("VerifyPassword(old) error = %v", err)
	}
	if oldValid {
		t.Fatal("VerifyPassword(old) = true, want false")
	}

	newValid, err := s.VerifyPassword("new-password")
	if err != nil {
		t.Fatalf("VerifyPassword(new) error = %v", err)
	}
	if !newValid {
		t.Fatal("VerifyPassword(new) = false, want true")
	}
}

func TestGenerateRandomPasswordReturnsRequestedLength(t *testing.T) {
	password, err := generateRandomPassword(24)
	if err != nil {
		t.Fatalf("generateRandomPassword() error = %v", err)
	}
	if len(password) != 24 {
		t.Fatalf("len(password) = %d, want 24", len(password))
	}
}
