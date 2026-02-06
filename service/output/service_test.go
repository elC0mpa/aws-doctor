package output

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/mocks/renderers"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name           string
		inputFormat    string
		expectedFormat Format
	}{
		{
			name:           "json format",
			inputFormat:    "json",
			expectedFormat: FormatJSON,
		},
		{
			name:           "table format explicit",
			inputFormat:    "table",
			expectedFormat: FormatTable,
		},
		{
			name:           "empty string defaults to table",
			inputFormat:    "",
			expectedFormat: FormatTable,
		},
		{
			name:           "unknown format defaults to table",
			inputFormat:    "unknown",
			expectedFormat: FormatTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.inputFormat)

			// Type assert to access internal format field
			s, ok := svc.(*service)
			if !ok {
				t.Fatal("NewService did not return *service type")
			}

			if s.format != tt.expectedFormat {
				t.Errorf("expected format %q, got %q", tt.expectedFormat, s.format)
			}

			if s.renderer == nil {
				t.Error("renderer should not be nil")
			}
		})
	}
}

func TestRenderCostComparison(t *testing.T) {
	t.Run("TableFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatTable, renderer: mr}

		input := model.RenderCostComparisonInput{
			AccountID:        "123",
			LastTotalCost:    "100.00 USD",
			CurrentTotalCost: "120.00 USD",
			LastMonth:        &model.CostInfo{},
			CurrentMonth:     &model.CostInfo{},
		}
		mr.On("DrawCostTable", input).Return()

		err := s.RenderCostComparison(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("JSONFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatJSON, renderer: mr}

		input := model.RenderCostComparisonInput{
			AccountID:        "123",
			LastTotalCost:    "100.00 USD",
			CurrentTotalCost: "120.00 USD",
			LastMonth:        &model.CostInfo{},
			CurrentMonth:     &model.CostInfo{},
		}
		mr.On("OutputCostComparisonJSON", input).Return(nil)

		err := s.RenderCostComparison(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})
}

func TestRenderTrend(t *testing.T) {
	t.Run("TableFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatTable, renderer: mr}

		mr.On("DrawTrendChart", "123", mock.Anything).Return()

		err := s.RenderTrend("123", []model.CostInfo{})
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("JSONFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatJSON, renderer: mr}

		mr.On("OutputTrendJSON", "123", mock.Anything).Return(nil)

		err := s.RenderTrend("123", []model.CostInfo{})
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})
}

func TestRenderWaste(t *testing.T) {
	t.Run("TableFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatTable, renderer: mr}

		input := model.RenderWasteInput{AccountID: "123"}
		mr.On("DrawWasteTable", input).Return()

		err := s.RenderWaste(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("JSONFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatJSON, renderer: mr}

		input := model.RenderWasteInput{AccountID: "123"}
		mr.On("OutputWasteJSON", input).Return(nil)

		err := s.RenderWaste(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})
}

func TestStopSpinner(t *testing.T) {
	mr := new(renderers.MockRenderer)
	s := &service{renderer: mr}

	mr.On("StopSpinner").Return()

	s.StopSpinner()
	mr.AssertExpectations(t)
}

func TestFormatConstants(t *testing.T) {
	if FormatTable != "table" {
		t.Errorf("FormatTable should be 'table', got %q", FormatTable)
	}

	if FormatJSON != "json" {
		t.Errorf("FormatJSON should be 'json', got %q", FormatJSON)
	}
}
