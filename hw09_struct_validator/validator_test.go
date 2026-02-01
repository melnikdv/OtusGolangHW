package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		in          interface{}
		expectedErr error
	}{
		{
			name: "valid User",
			in: User{
				ID:    "123e4567-e89b-12d3-a456-426614174000",
				Age:   25,
				Email: "test@test.com",
				Role:  "admin",
				Phones: []string{
					"79991234567",
				},
			},
			expectedErr: nil,
		},
		{
			name: "User with multiple errors",
			in: User{
				ID:    "short",
				Age:   10,
				Email: "invalid-email",
				Role:  "guest",
				Phones: []string{
					"123",
					"abc",
				},
			},
			expectedErr: ValidationErrors{}, // ожидаем ValidationErrors
		},
		{
			name: "Response with invalid code",
			in: Response{
				Code: 418,
				Body: "ok",
			},
			expectedErr: ValidationErrors{}, // ожидаем ValidationErrors
		},
		{
			name:        "non-struct input",
			in:          42,
			expectedErr: ErrNotStruct,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d: %s", i, tt.name), func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

			// Проверяем program errors
			if errors.Is(tt.expectedErr, ErrNotStruct) {
				if !errors.Is(err, ErrNotStruct) {
					t.Fatalf("expected ErrNotStruct, got %v", err)
				}
				return
			}

			// Проверяем ValidationErrors
			var vErrs ValidationErrors
			if !errors.As(err, &vErrs) {
				t.Fatalf("expected ValidationErrors, got %T", err)
			}

			if len(vErrs) == 0 {
				t.Fatal("expected at least one validation error")
			}
		})
	}
}

// Проверка накопления ошибок.
func TestValidate_MultipleErrors(t *testing.T) {
	u := User{
		ID:    "short-id",
		Age:   10,
		Email: "invalid-email",
		Role:  "guest",
		Phones: []string{
			"123",
			"abc",
		},
	}

	err := Validate(u)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var vErrs ValidationErrors
	if !errors.As(err, &vErrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	if len(vErrs) < 4 {
		t.Fatalf("expected multiple validation errors, got %d", len(vErrs))
	}
}

// Проверка errors.Is.
func TestValidate_ErrorsIs(t *testing.T) {
	u := User{
		Age: 10, // min:18
	}

	err := Validate(u)
	if err == nil {
		t.Fatal("expected error")
	}

	var vErrs ValidationErrors
	_ = errors.As(err, &vErrs)

	found := false
	for _, e := range vErrs {
		if errors.Is(e.Err, ErrMin) {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected ErrMin validation error")
	}
}

// Валидация слайса.
func TestValidate_Slice(t *testing.T) {
	r := Response{
		Code: 418,
	}

	err := Validate(r)
	if err == nil {
		t.Fatal("expected validation error")
	}

	var vErrs ValidationErrors
	if !errors.As(err, &vErrs) {
		t.Fatalf("expected ValidationErrors")
	}

	if len(vErrs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(vErrs))
	}

	if !errors.Is(vErrs[0].Err, ErrIn) {
		t.Fatalf("expected ErrIn, got %v", vErrs[0].Err)
	}
}

// Не структура — программная ошибка.
func TestValidate_NotStruct(t *testing.T) {
	err := Validate(42)
	if !errors.Is(err, ErrNotStruct) {
		t.Fatalf("expected ErrNotStruct, got %v", err)
	}
}
