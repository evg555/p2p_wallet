package repository

import (
	"context"

	"p2p_wallet/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracingWalletRepo struct {
	next   WalletRepo
	tracer trace.Tracer
}

func NewWalletRepoWithTracing(next WalletRepo) WalletRepo {
	return &tracingWalletRepo{
		next:   next,
		tracer: otel.Tracer("p2p-wallet/repository/wallet"),
	}
}

func (r *tracingWalletRepo) Save(ctx context.Context, wallet *domain.Wallet) (*domain.Wallet, error) {
	ctx, span := r.tracer.Start(ctx, "repo.wallet.save", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("repo.method", "Save"),
	)

	res, err := r.next.Save(ctx, wallet)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errorClass(err))
	}
	span.End()

	return res, err
}

func (r *tracingWalletRepo) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Wallet, error) {
	ctx, span := r.tracer.Start(ctx, "repo.wallet.find_by_user_id", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("repo.method", "FindByUserID"),
	)

	res, err := r.next.FindByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errorClass(err))
	}
	span.End()

	return res, err
}

func (r *tracingWalletRepo) FindByID(ctx context.Context, id domain.WalletID) (*domain.Wallet, error) {
	ctx, span := r.tracer.Start(ctx, "repo.wallet.find_by_id", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("repo.method", "FindByID"),
	)

	res, err := r.next.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errorClass(err))
	}
	span.End()

	return res, err
}
