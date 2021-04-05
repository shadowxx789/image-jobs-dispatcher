package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_Parse(t *testing.T) {
	authService := NewService(Opts{})
	tbl := []struct {
		c   string
		res *Claims
	}{
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJ0aWQiOjEsIm9pZCI6MSwiYXVkIjoiY29tLmNvbXBhbnkuam9ic2VydmljZSIsImF6cCI6IjEiLCJlbWFpbCI6ImN1c3RvbWVyQG1haWwuY29tIn0.CcTapGbWX0UEMovUwC8kAcWMUxmbOeO0qhsu-wqHQH0",
			&Claims{StandardClaims:jwt.StandardClaims{Audience:"com.company.jobservice", ExpiresAt:0, Id:"", IssuedAt:1516239022, Issuer:"", NotBefore:0, Subject:"1234567890"}, TenantID:1, ClientID:1, AppID:"1", Name:"John Doe", Email:"customer@mail.com"},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyMjIyMjIiLCJuYW1lIjoiQWxleCBBbGV4IiwiaWF0IjoxMTExMTExMSwidGlkIjoyLCJvaWQiOjIsImF1ZCI6ImNvbS5uby5jb21wYW55IiwiYXpwIjoiMiIsImVtYWlsIjoibmV3QG1haWwuY29tIn0.beHh_d3mDO4ufQBySzukbYE9cDvSZ6KRXDFfBlnHq_s",
			&Claims{StandardClaims:jwt.StandardClaims{Audience:"com.no.company", ExpiresAt:0, Id:"", IssuedAt:11111111, Issuer:"", NotBefore:0, Subject:"222222"}, TenantID:2, ClientID:2, AppID:"2", Name:"Alex Alex", Email:"new@mail.com"},
		},
		{
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyMjIyMjIiLCJuYW1lIjoiQWxleCBBbGV4IiwiaWF0IjoxMTExMTExMSwidGlkIjoyLCJvaWQiOjIsImF1ZCI6ImNvbS5uby5jb21wYW55IiwiYXpwIjoiMiIsImVtYWlsIjoibmV3QG1haWwuY29tIn0.WK9Ex-3x08E1dFMUJphxhbII2mjtuNVt5DUPDhhymag",
			&Claims{StandardClaims:jwt.StandardClaims{Audience:"com.no.company", ExpiresAt:0, Id:"", IssuedAt:11111111, Issuer:"", NotBefore:0, Subject:"222222"}, TenantID:2, ClientID:2, AppID:"2", Name:"Alex Alex", Email:"new@mail.com"},
		},
		{
			"pwIjoiMiIsImVtYWlsIjoibmV3QG1haWwuY29tIn0.beHh_d3mDO4ufQBySzukbYE9cDvSZ6KRXDFfBlnHq_s",
			&Claims{StandardClaims:jwt.StandardClaims{Audience:"com.no.company", ExpiresAt:0, Id:"", IssuedAt:11111111, Issuer:"", NotBefore:0, Subject:"222222"}, TenantID:2, ClientID:2, AppID:"2", Name:"Alex Alex", Email:"new@mail.com"},

		},
	}
	for i, tt := range tbl {
		claims, err := authService.Parse(tt.c)
		if err != nil {
			t.Logf("test case with error #%d\n", i)
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, claims, tt.res, "test case #%d", i)
	}
}