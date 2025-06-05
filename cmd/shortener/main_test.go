package main

import (
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestServerCreation проверяет корректность создания сервера
func TestServerCreation(t *testing.T) {
	var s app.Server
	server := s.NewServer()
	router := server.Router
	handler := server.Handler

	require.NotNilf(t, server, "Expected server instance, got nil")

	require.NotNilf(t, router, "Expected router to be initialized, got nil")

	require.NotNilf(t, handler, "Expected handler to be initialized, got nil")
}
