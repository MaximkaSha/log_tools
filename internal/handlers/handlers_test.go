package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/go-chi/chi/v5"
)

func TestHandlers_HandleUpdate(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name   string
		want   want
		url    string
		method string
	}{
		// определяем все тесты
		{
			name: "positive counter #1",
			want: want{
				code:        200,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/update/counter/TestCount/100",
			method: "POST",
		},
		{
			name: "positive gauge #2",
			want: want{
				code:        200,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/update/gauge/TestCount/100.000",
			method: "POST",
		},
		{
			name: "negative type mismatch",
			want: want{
				code:        501,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/update/gOuge/TestCount/100",
			method: "POST",
		},
		{
			name: "negative bad data",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/update/gauge/TestCount/bad_data",
			method: "POST",
		},
		{
			name: "negative no data",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/update/",
			method: "POST",
		},
		{
			name: "negative data error",
			want: want{
				code:        501,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/update/gouge/TestCount/misstype",
			method: "POST",
		},
		{
			name: "get value possitive", //надо значение из body проверить
			want: want{
				code:        200,
				contentType: "",
			},
			url:    "/value/gauge/TestCount",
			method: "GET",
		},
		{
			name: "get negative", //надо значение из body проверить
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/value/gauge/TestCount123",
			method: "GET",
		},
		{
			name: "negative type mismatch",
			want: want{
				code:        501,
				contentType: "text/plain; charset=utf-8",
			},
			url:    "/value/gOuge/TestCount",
			method: "GET",
		},
		{
			name: "get home", //надо значение из body проверить
			want: want{
				code:        200,
				contentType: "",
			},
			url:    "/",
			method: "home",
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			var request = new(http.Request)
			if tt.method == "POST" {
				request = httptest.NewRequest(http.MethodPost, tt.url, nil)
			}
			if tt.method == "GET" || tt.method == "home" {
				request = httptest.NewRequest(http.MethodGet, tt.url, nil)
			}
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			repo := storage.NewRepo()
			handl := NewHandlers(repo)
			mux := chi.NewRouter()
			if tt.method == "POST" {
				mux.Post("/update/{type}/{name}/{value}", handl.HandleUpdate)
			}
			if tt.method == "GET" {
				handl.repo.InsertData("gauge", "TestCount", "100.00")
				mux.Get("/value/{type}/{name}", handl.HandleGetUpdate)
			}
			if tt.method == "home" {

				handl.repo.InsertData("gauge", "TestCount", "100.00")
				mux.Get("/", handl.HandleGetHome)
			}
			// запускаем сервер
			mux.ServeHTTP(w, request)
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

func TestHandlers_HandleGetUpdate(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		obj  Handlers
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.obj.HandleGetUpdate(tt.args.w, tt.args.r)
		})
	}
}
