package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaximkaSha/log_tools/internal/storage"
)

func TestHandlers_HandleUpdate(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		want want
		url  string
	}{
		// определяем все тесты
		{
			name: "positive counter #1",
			want: want{
				code:        200,
				contentType: "text/plain; charset=utf-8",
			},
			url: "/update/counter/TestCount/100",
		},
		{
			name: "positive gauge #2",
			want: want{
				code:        200,
				contentType: "text/plain; charset=utf-8",
			},
			url: "/update/gauge/TestCount/100",
		},
		{
			name: "negative type mismatch",
			want: want{
				code:        501,
				contentType: "text/plain; charset=utf-8",
			},
			url: "/update/gouge/TestCount/100",
		},
		{
			name: "negative no data",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
			},
			url: "/update/",
		},
		{
			name: "negative data error",
			want: want{
				code:        501,
				contentType: "text/plain; charset=utf-8",
			},
			url: "/update/gouge/TestCount/misstype",
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.url, nil)
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			repo := storage.NewRepo()
			handl := NewHandlers(repo)
			h := http.HandlerFunc(handl.HandleUpdate)
			// запускаем сервер
			h.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			// заголовок ответа
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
