package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeWhenOK(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
		fmt.Println(response.Body.String())
	}
}

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}

	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"
	totalCafe := len(cafeList[city])

	requests := []struct {
		count int // передаваемое значение count
		want  int // ожидаемое количество кафе в ответе
	}{
		{count: 0, want: 0},
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: totalCafe},
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)
		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")

		var actualCount int
		if body == "" {
			actualCount = 0
		} else {
			nonEmpty := make([]string, 0)
			for _, c := range cafes {
				if strings.TrimSpace(c) != "" {
					nonEmpty = append(nonEmpty, c)
				}
			}
			actualCount = len(nonEmpty)
		}
		assert.Equal(t, v.want, actualCount)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&search=%s", city, v.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, url, nil)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")

		var realCafes []string
		for _, c := range cafes {
			if strings.TrimSpace(c) != "" {
				realCafes = append(realCafes, c)
			}
		}

		if v.wantCount > 0 {
			lowerSearch := strings.ToLower(v.search)
			for _, cafe := range realCafes {
				lowerCafe := strings.ToLower(cafe)
				assert.True(t, strings.Contains(lowerCafe, lowerSearch))
			}
		}
	}
}
