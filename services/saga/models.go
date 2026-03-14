package saga

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SagaStatus string

const (
	SagaStatusRunning      SagaStatus = "RUNNING"
	SagaStatusCompleted    SagaStatus = "COMPLETED"
	SagaStatusCompensating SagaStatus = "COMPENSATING"
	SagaStatusCompensated  SagaStatus = "COMPENSATED"
	SagaStatusFailed       SagaStatus = "FAILED"
)

type StepStatus string

const (
	StepStatusPending     StepStatus = "PENDING"
	StepStatusRunning     StepStatus = "RUNNING"
	StepStatusCompleted   StepStatus = "COMPLETED"
	StepStatusCompensated StepStatus = "COMPENSATED"
	StepStatusFailed      StepStatus = "FAILED"
)

type Saga struct {
	ID                uuid.UUID  `json:"id"`
	Kind              string     `json:"kind"`
	Status            SagaStatus `json:"status"`
	CurrentStep       int        `json:"current_step"`
	Input             []byte     `json:"input"`
	Output            []byte     `json:"output,omitempty"`
	CompensationState []byte     `json:"compensation_state,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type SagaStep struct {
	ID          uuid.UUID  `json:"id"`
	SagaID      uuid.UUID  `json:"saga_id"`
	Name        string     `json:"name"`
	Seq         int        `json:"seq"`
	Status      StepStatus `json:"status"`
	Input       []byte     `json:"input,omitempty"`
	Output      []byte     `json:"output,omitempty"`
	Error       string     `json:"error,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type SagaDefinition struct {
	Name          string
	Steps         []StepDefinition
	Compensations []StepDefinition
}

type StepDefinition struct {
	Name       string
	Action     func(ctx context.Context, input []byte) ([]byte, error)
	Compensate func(ctx context.Context, output []byte) error
}

type Repository interface {
	CreateSaga(ctx context.Context, saga *Saga) error
	GetSaga(ctx context.Context, id uuid.UUID) (*Saga, error)
	UpdateSaga(ctx context.Context, saga *Saga) error
	GetSagasByStatus(ctx context.Context, status SagaStatus) ([]Saga, error)

	CreateSagaStep(ctx context.Context, step *SagaStep) error
	GetSagaSteps(ctx context.Context, sagaID uuid.UUID) ([]SagaStep, error)
	UpdateSagaStep(ctx context.Context, step *SagaStep) error
}
