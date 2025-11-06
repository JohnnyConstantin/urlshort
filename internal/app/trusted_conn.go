package app

import (
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"net"
	"net/http"
	"strings"
)

// isIPInTrustedSubnet проверяет, принадлежит ли IP доверенной подсети
func isIPInTrustedSubnet(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, trustedNet, err := net.ParseCIDR(config.Options.TrustedSubnet) // Использую встроенную функцию net
	if err != nil {
		return false
	}

	return trustedNet.Contains(ip)
}

// RequireTrustedIP Middleware для проверки доверенного IP
func RequireTrustedIP(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if config.Options.TrustedSubnet == "" {
			http.Error(w, "This endpoint is forbidden", http.StatusForbidden)
			return
		}
		// Получаем IP из заголовка X-Real-IP
		realIP := r.Header.Get("X-Real-IP")
		if realIP == "" {
			http.Error(w, "X-Real-IP header required", http.StatusForbidden)
			return
		}

		// Санитайзим IP (хотя бы минимально)
		cleanIP := strings.Split(realIP, ":")[0]

		// Проверяем принадлежность к доверенной подсети
		if !isIPInTrustedSubnet(cleanIP) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		h(w, r.WithContext(r.Context()))
	}
}
