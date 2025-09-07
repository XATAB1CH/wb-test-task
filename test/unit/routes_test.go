package unit

import (
	"context"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"wb-test-task/internal/models"
	"wb-test-task/internal/routes"
	"wb-test-task/internal/service"
)

type noopRepo struct{}

func (n *noopRepo) SaveOrder(ctx context.Context, o models.Order) error {
	return nil
}

func (n *noopRepo) GetOrder(ctx context.Context, uid string) (*models.Order, error) {
	return nil, context.Canceled
}

type noopCache struct{}

func (n *noopCache) Get(key string) (*models.Order, bool) {
	return nil, false
}

func (n *noopCache) Set(key string, value *models.Order) {
	return
}
func (n *noopCache) Delete(key string) {
	return
}

func TestRoutes_IndexServed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpl := template.Must(template.New("index.html").Parse(`OK`))
	r.SetHTMLTemplate(tmpl)

	svc := service.NewOrderService(&noopRepo{}, &noopCache{})
	r = routes.InitRoutes(r, svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "OK", w.Body.String())
}
