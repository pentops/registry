package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/pentops/go-grpc-helpers/grpcerror"
	"github.com/pentops/go-grpc-helpers/protovalidatemw"
	"github.com/pentops/log.go/grpc_log"
	"github.com/pentops/log.go/log"
	"github.com/pentops/o5-go/auth/v1/auth_pb"
	"github.com/pentops/protostate/gen/state/v1/psm_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCMiddleware() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpc_log.UnaryServerInterceptor(log.DefaultContext, log.DefaultTrace, log.DefaultLogger),
		grpcerror.UnaryServerInterceptor(log.DefaultLogger),
		PSMActionMiddleware(actorExtractor),
		protovalidatemw.UnaryServerInterceptor(),
	}
}

type DBConfig struct {
	URL string `env:"POSTGRES_URL"`
}

func (cfg *DBConfig) OpenDatabase(ctx context.Context) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, err
	}

	// Default is unlimited connections, use a cap to prevent hammering the database if it's the bottleneck.
	// 10 was selected as a conservative number and will likely be revised later.
	db.SetMaxOpenConns(10)

	for {
		if err := db.Ping(); err != nil {
			log.WithError(ctx, err).Error("pinging PG")
			time.Sleep(time.Second)
			continue
		}
		break
	}

	return db, nil
}

func actorExtractor(ctx context.Context) *auth_pb.Actor {
	return &auth_pb.Actor{
		Type: &auth_pb.Actor_Named{
			Named: &auth_pb.Actor_NamedActor{
				Name: "Unauthenticated Client",
			},
		},
	}
}

type actionContextKey struct{}

type PSMAction struct {
	Method string
	Actor  *auth_pb.Actor
}

// PSMCause is a gRPC middleware that injects the PSM cause into the context.
func PSMActionMiddleware(actorExtractor func(context.Context) *auth_pb.Actor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		actor := actorExtractor(ctx)
		cause := PSMAction{
			Method: info.FullMethod,
			Actor:  actor,
		}
		ctx = context.WithValue(ctx, actionContextKey{}, cause)
		return handler(ctx, req)
	}
}

func WithPSMAction(ctx context.Context, action PSMAction) context.Context {
	return context.WithValue(ctx, actionContextKey{}, action)
}

func CommandCause(ctx context.Context) (*psm_pb.Cause, error) {

	cause, ok := ctx.Value(actionContextKey{}).(PSMAction)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no actor")
	}

	return &psm_pb.Cause{
		Type: &psm_pb.Cause_Command{
			Command: &psm_pb.CommandCause{
				MethodName: cause.Method,
				Actor:      cause.Actor,
			},
		},
	}, nil
}
