package repository

import (
	"context"
	"fmt"

	"p2p_wallet/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracingUserRepo struct {
	next   UserRepo
	tracer trace.Tracer
}

func NewUserRepoWithTracing(next UserRepo) UserRepo {
	return &tracingUserRepo{
		next:   next,
		tracer: otel.Tracer("p2p-wallet/repository/user"),
	}
}

func (r *tracingUserRepo) Save(ctx context.Context, user *domain.User) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "repo.user.save", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("repo.method", "Save"),
	)

	res, err := r.next.Save(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errorClass(err))
	}
	span.End()

	return res, err
}

func (r *tracingUserRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "repo.user.get_by_login", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("repo.method", "GetByLogin"),
	)

	res, err := r.next.GetByLogin(ctx, login)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errorClass(err))
	}
	span.End()

	return res, err
}

func (r *tracingUserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "repo.user.get_by_id", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("repo.method", "GetByID"),
		attribute.String("user.id", fmt.Sprintf("%d", id)),
	)

	res, err := r.next.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errorClass(err))
	}
	span.End()

	return res, err
}
