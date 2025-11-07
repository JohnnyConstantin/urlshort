package grpcserver

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

func GRPCLoggingInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		logger.Infow("gRPC request started",
			"method", info.FullMethod,
			"request", req,
		)

		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			logger.Errorw("gRPC request failed",
				"method", info.FullMethod, // Вместо эндпоинта логируем дернутый метод
				"duration", duration, // Аналогично с HTTP - логируем время выполнения
				"error", err,
			)
		} else {
			logger.Infow("gRPC request completed",
				"method", info.FullMethod, // Вместо эндпоинта логируем дернутый метод
				"duration", duration, // Аналогично с HTTP - логируем время выполнения
			)
		}

		return resp, err
	}
}

func GRPCAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Пропускаем аутентификацию для публичных методов
	if isPublicMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimPrefix(authHeaders[0], "Bearer ")
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	userID, err := uuid.Parse(token) // Пользователь должен передать в хедере user id в виде UUID.
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid user id")
	}

	ctx = context.WithValue(ctx, userIDKey, userID)

	return handler(ctx, req)
}

func isPublicMethod(fullMethod string) bool {
	publicMethods := map[string]bool{
		"/shortener.Shortener/Ping": true, // Хотя бы один для демонстрации
	}
	return publicMethods[fullMethod]
}

// Получить userID из контекста
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID.String(), ok
}
