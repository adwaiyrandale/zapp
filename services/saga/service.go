package saga

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSagaNotRunning      = errors.New("saga is not running")
	ErrStepAlreadyExecuted = errors.New("step already executed")
	ErrNoMoreSteps         = errors.New("no more steps")
	ErrSagaFailed          = errors.New("saga execution failed")
)

type Service struct {
	repo        Repository
	definitions map[string]SagaDefinition
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:        repo,
		definitions: make(map[string]SagaDefinition),
	}
}

func (s *Service) RegisterSaga(def SagaDefinition) {
	s.definitions[def.Name] = def
}

type StartSagaInput struct {
	Kind  string
	Input interface{}
}

func (s *Service) StartSaga(ctx context.Context, input *StartSagaInput) (*Saga, error) {
	def, ok := s.definitions[input.Kind]
	if !ok {
		return nil, errors.New("unknown saga kind: " + input.Kind)
	}

	inputBytes, err := json.Marshal(input.Input)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	saga := &Saga{
		ID:          uuid.New(),
		Kind:        input.Kind,
		Status:      SagaStatusRunning,
		CurrentStep: 0,
		Input:       inputBytes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateSaga(ctx, saga); err != nil {
		return nil, err
	}

	for i, stepDef := range def.Steps {
		step := &SagaStep{
			ID:     uuid.New(),
			SagaID: saga.ID,
			Name:   stepDef.Name,
			Seq:    i,
			Status: StepStatusPending,
		}
		if err := s.repo.CreateSagaStep(ctx, step); err != nil {
			return nil, err
		}
	}

	return saga, nil
}

func (s *Service) GetSaga(ctx context.Context, id uuid.UUID) (*Saga, error) {
	return s.repo.GetSaga(ctx, id)
}

func (s *Service) GetSagaSteps(ctx context.Context, sagaID uuid.UUID) ([]*SagaStep, error) {
	steps, err := s.repo.GetSagaSteps(ctx, sagaID)
	if err != nil {
		return nil, err
	}
	result := make([]*SagaStep, len(steps))
	for i := range steps {
		result[i] = &steps[i]
	}
	return result, nil
}

func (s *Service) ExecuteNextStep(ctx context.Context, sagaID uuid.UUID) (*Saga, *SagaStep, error) {
	saga, err := s.repo.GetSaga(ctx, sagaID)
	if err != nil {
		return nil, nil, err
	}

	if saga.Status != SagaStatusRunning {
		return nil, nil, ErrSagaNotRunning
	}

	def, ok := s.definitions[saga.Kind]
	if !ok {
		return nil, nil, errors.New("unknown saga kind")
	}

	if saga.CurrentStep >= len(def.Steps) {
		return nil, nil, ErrNoMoreSteps
	}

	stepDef := def.Steps[saga.CurrentStep]
	steps, err := s.repo.GetSagaSteps(ctx, sagaID)
	if err != nil {
		return nil, nil, err
	}

	currentStep := &steps[saga.CurrentStep]
	if currentStep.Status != StepStatusPending {
		return nil, nil, ErrStepAlreadyExecuted
	}

	now := time.Now().UTC()
	currentStep.Status = StepStatusRunning
	currentStep.StartedAt = &now
	if err := s.repo.UpdateSagaStep(ctx, currentStep); err != nil {
		return nil, nil, err
	}

	var stepInput []byte
	if saga.CurrentStep == 0 {
		stepInput = saga.Input
	} else {
		prevStep := &steps[saga.CurrentStep-1]
		stepInput = prevStep.Output
	}

	output, err := stepDef.Action(ctx, stepInput)
	if err != nil {
		currentStep.Status = StepStatusFailed
		currentStep.Error = err.Error()
		now := time.Now().UTC()
		currentStep.CompletedAt = &now
		s.repo.UpdateSagaStep(ctx, currentStep)

		saga.Status = SagaStatusCompensating
		s.repo.UpdateSaga(ctx, saga)

		return saga, currentStep, err
	}

	currentStep.Status = StepStatusCompleted
	currentStep.Output = output
	now = time.Now().UTC()
	currentStep.CompletedAt = &now
	s.repo.UpdateSagaStep(ctx, currentStep)

	saga.CurrentStep++
	if saga.CurrentStep >= len(def.Steps) {
		saga.Status = SagaStatusCompleted
		now = time.Now().UTC()
		saga.CompletedAt = &now
	}
	saga.Output = output
	s.repo.UpdateSaga(ctx, saga)

	return saga, currentStep, nil
}

func (s *Service) Compensate(ctx context.Context, sagaID uuid.UUID) (*Saga, error) {
	saga, err := s.repo.GetSaga(ctx, sagaID)
	if err != nil {
		return nil, err
	}

	def, ok := s.definitions[saga.Kind]
	if !ok {
		return nil, errors.New("unknown saga kind")
	}

	if saga.Status != SagaStatusCompensating {
		return nil, errors.New("saga is not in compensating state")
	}

	steps, err := s.repo.GetSagaSteps(ctx, sagaID)
	if err != nil {
		return nil, err
	}

	for i := saga.CurrentStep - 1; i >= 0; i-- {
		step := &steps[i]
		if step.Status != StepStatusCompleted {
			continue
		}

		if i < len(def.Compensations) {
			compDef := def.Compensations[i]
			err := compDef.Compensate(ctx, step.Output)
			if err != nil {
				step.Status = StepStatusFailed
				step.Error = err.Error()
				s.repo.UpdateSagaStep(ctx, step)
				continue
			}
		}

		step.Status = StepStatusCompensated
		now := time.Now().UTC()
		step.CompletedAt = &now
		s.repo.UpdateSagaStep(ctx, step)
	}

	saga.Status = SagaStatusCompensated
	now := time.Now().UTC()
	saga.CompletedAt = &now
	s.repo.UpdateSaga(ctx, saga)

	return saga, nil
}

func (s *Service) GetRunningSagas(ctx context.Context) ([]Saga, error) {
	return s.repo.GetSagasByStatus(ctx, SagaStatusRunning)
}

func (s *Service) GetCompensatingSagas(ctx context.Context) ([]Saga, error) {
	return s.repo.GetSagasByStatus(ctx, SagaStatusCompensating)
}
