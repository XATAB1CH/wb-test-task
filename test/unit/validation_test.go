package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wb-test-task/internal/validation"
)

type currencyDTO struct {
	Currency string `validate:"required,iso4217"`
}

func TestISO4217_ValidCodes(t *testing.T) {
	v := currencyDTO{Currency: "RUB"}
	err := validation.ValidateStruct(v)
	require.NoError(t, err)

	v = currencyDTO{Currency: "USD"}
	err = validation.ValidateStruct(v)
	require.NoError(t, err)

	v = currencyDTO{Currency: "EUR"}
	err = validation.ValidateStruct(v)
	require.NoError(t, err)
}

type userDTO struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

type requiredDTO struct {
	Title   string `validate:"required"`
	Count   int    `validate:"required"`
	Enabled bool   `validate:"required"`
}

func TestEmailValidation_Valid(t *testing.T) {
	cases := []userDTO{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob.smith+dev@sub.domain.co"},
		{Name: "Юзер", Email: "user.name@пример.рф"},
	}

	for _, c := range cases {
		err := validation.ValidateStruct(c)
		require.NoError(t, err)
	}
}

func TestEmailValidation_Invalid(t *testing.T) {
	cases := []userDTO{
		{Name: "", Email: "user@example.com"},
		{Name: "NoAt", Email: "no-at.example.com"},
		{Name: "NoDomain", Email: "user@"},
		{Name: "Spaces", Email: "user name@example.com"},
		{Name: "Empty", Email: ""},
	}

	for _, c := range cases {
		err := validation.ValidateStruct(c)
		assert.Error(t, err)
	}
}

func TestRequiredFields(t *testing.T) {
	ok := requiredDTO{Title: "ok", Count: 1, Enabled: true}
	require.NoError(t, validation.ValidateStruct(ok))

	err := validation.ValidateStruct(requiredDTO{Title: "", Count: 1, Enabled: true})
	assert.Error(t, err)

	err = validation.ValidateStruct(requiredDTO{Title: "x", Count: 0, Enabled: true})
	assert.Error(t, err)

	err = validation.ValidateStruct(requiredDTO{Title: "x", Count: 2, Enabled: false})
	assert.Error(t, err)
}
