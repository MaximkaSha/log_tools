package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				contentType: "text/html",
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
				contentType: "text/html",
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
				handl.Repo.InsertData("gauge", "TestCount", "100.00")
				mux.Get("/value/{type}/{name}", handl.HandleGetUpdate)
			}
			if tt.method == "home" {

				handl.Repo.InsertData("gauge", "TestCount", "100.00")
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

func TestHandlers_HandlePostJSONUpdate(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		contentType string
		body        string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name        string
		want        want
		url         string
		method      string
		contentType string
		data        string
	}{
		// определяем все тесты
		{
			name: "positive json #1",
			want: want{
				code:        200,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"gauge","value":1072448}`,
			},
			url:         "/update/",
			method:      "POST",
			data:        `{"id":"Alloc","type":"gauge","value":1072448}`,
			contentType: "application/json",
		},
		{
			name: "positive json #2",
			want: want{
				code:        200,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"gauge","value":1072448.001}`,
			},
			url:         "/update/",
			method:      "POST",
			data:        `{"id":"Alloc","type":"gauge","value":1072448.001}`,
			contentType: "application/json",
		},
		{
			name: "positive json #3",
			want: want{
				code:        200,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"counter","value":1072448}`,
			},
			url:         "/update/",
			method:      "POST",
			data:        `{"id":"Alloc","type":"counter","value":1072448}`,
			contentType: "application/json",
		},
		{
			name: "negative json #4",
			want: want{
				code:        404,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"counter","value":1072448}`,
			},
			url:         "/update/",
			method:      "POST",
			data:        ``,
			contentType: "text/html",
		},
		{
			name: "negative json #5",
			want: want{
				code:        404,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"counter","value":1072448.0001}`,
			},
			url:         "/update/",
			method:      "POST",
			data:        ``,
			contentType: "application/json",
		},
		{
			name: "negative json #6",
			want: want{
				code:        404,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"gauge","value":sdfghgfd}`,
			},
			url:         "/update/",
			method:      "POST",
			data:        ``,
			contentType: "application/json",
		},
		{
			name: "negative json #7",
			want: want{
				code:        404,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"asdfgfd","value":1072448.0001}`,
			},
			url:         "/update/",
			method:      "POST",
			data:        ``,
			contentType: "application/json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := storage.NewRepo()
			var request = new(http.Request)
			//data := strings.NewReader(tt.data)
			//data, _ := json.Marshal(tt.data)
			request = httptest.NewRequest(http.MethodPost, tt.url, strings.NewReader(tt.data))
			request.Header.Add("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			srv, _ := NewTestServer(&repo)
			srv.ServeHTTP(w, request)
			resp := w.Result()
			respBody, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				t.Fail()
			}

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			if tt.data != `` {
				require.JSONEq(t, tt.want.body, string(respBody))
			}

		})
	}
}

func TestHandlers_HandlePostJSONValue(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		contentType string
		body        string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name        string
		want        want
		url         string
		method      string
		contentType string
		data        string
	}{
		// определяем все тесты
		{
			name: "positive json #1",
			want: want{
				code:        200,
				contentType: "application/json",
				body:        `{"id":"Alloc","type":"gauge","value":1072448,"delta":0}`,
			},
			url:         "/value/",
			method:      "POST",
			data:        `{"id":"Alloc","type":"gauge"}`,
			contentType: "application/json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := storage.NewRepo()
			var request = new(http.Request)
			//data := strings.NewReader(tt.data)
			//data, _ := json.Marshal(tt.data)
			request = httptest.NewRequest(http.MethodPost, tt.url, strings.NewReader(tt.data))
			request.Header.Add("Content-Type", tt.contentType)
			w := httptest.NewRecorder()
			srv, handl := NewTestServer(&repo)
			var model models.Metrics
			json.Unmarshal([]byte(tt.want.body), &model)
			handl.Repo.InsertMetric(model)
			srv.ServeHTTP(w, request)
			resp := w.Result()
			respBody, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				t.Fail()
			}

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			if tt.data != `` {
				require.JSONEq(t, tt.want.body, string(respBody))
			}

		})
	}
}

func NewTestServer(repo *storage.Repository) (*chi.Mux, *Handlers) {
	handl := NewHandlers(*repo)
	mux := chi.NewRouter()
	mux.Post("/update/", handl.HandlePostJSONUpdate)
	mux.Post("/value/", handl.HandlePostJSONValue)

	return mux, &handl

}
