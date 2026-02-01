//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})

	// Новые тесты:

	t.Run("case insensitive domain matching", func(t *testing.T) {
		// Проверяем, что домен ищется без учёта регистра
		result, err := GetDomainStat(bytes.NewBufferString(data), "COM")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("empty input", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(""), "com")
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("input with empty lines", func(t *testing.T) {
		dataWithEmptyLines := "\n\n" + data + "\n\n"
		result, err := GetDomainStat(bytes.NewBufferString(dataWithEmptyLines), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("email without @", func(t *testing.T) {
		invalidData := `{"Id":1,"Email":"invalid-email"}`
		result, err := GetDomainStat(bytes.NewBufferString(invalidData), "com")
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("email with multiple @", func(t *testing.T) {
		// Согласно RFC, в email может быть только один @, но на практике...
		// Мы используем LastIndex, так что возьмём часть после последнего @
		multiAtData := `{"Id":1,"Email":"user@sub.domain@evil.com"}`
		result, err := GetDomainStat(bytes.NewBufferString(multiAtData), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"evil.com": 1}, result)
	})

	t.Run("exact suffix match only", func(t *testing.T) {
		// Убедимся, что "ocom" не совпадает с "com"
		trickyData := `{"Id":1,"Email":"test@fakeocom"}`
		result, err := GetDomainStat(bytes.NewBufferString(trickyData), "com")
		require.NoError(t, err)
		require.Empty(t, result)
	})

	t.Run("domain with subdomains", func(t *testing.T) {
		subdomainData := `{"Id":1,"Email":"user@mail.sub.example.com"}`
		result, err := GetDomainStat(bytes.NewBufferString(subdomainData), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"mail.sub.example.com": 1}, result)
	})

	t.Run("non-ASCII domains (IDN)", func(t *testing.T) {
		// Проверяем корректность работы с нижним регистром для кириллицы
		idnData := `{"Id":1,"Email":"user@ПРИМЕР.РФ"}`
		result, err := GetDomainStat(bytes.NewBufferString(idnData), "рф")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"пример.рф": 1}, result)
	})
}
