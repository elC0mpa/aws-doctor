package output

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/mocks/renderers"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
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
			name:           "csv format",
			inputFormat:    "csv",
			expectedFormat: FormatCSV,
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
	input := model.RenderCostComparisonInput{
		AccountID:        "123",
		LastTotalCost:    "100.00 USD",
		CurrentTotalCost: "120.00 USD",
		LastMonth:        &model.CostInfo{},
		CurrentMonth:     &model.CostInfo{},
	}

	t.Run("TableFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatTable, renderer: mr}
		mr.On("DrawCostTable", input).Return()

		err := s.RenderCostComparison(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("JSONFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatJSON, renderer: mr}
		mr.On("OutputCostComparisonJSON", input).Return(nil)

		err := s.RenderCostComparison(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("CSVFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatCSV, renderer: mr}
		mr.On("OutputCostComparisonCSV", input).Return(nil)

		err := s.RenderCostComparison(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})
}

func TestRenderTrend(t *testing.T) {
	costs := []model.CostInfo{}
	services := []string{"redshift"}

	t.Run("TableFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatTable, renderer: mr}
		mr.On("DrawTrendChart", "123", costs).Return()

		err := s.RenderTrend("123", costs, services)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("JSONFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatJSON, renderer: mr}
		mr.On("OutputTrendJSON", "123", costs, services).Return(nil)

		err := s.RenderTrend("123", costs, services)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("CSVFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatCSV, renderer: mr}
		mr.On("OutputTrendCSV", costs, services).Return(nil)

		err := s.RenderTrend("123", costs, services)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})
}

func TestRenderWaste(t *testing.T) {
	input := model.RenderWasteInput{AccountID: "123"}

	t.Run("TableFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatTable, renderer: mr}
		mr.On("DrawWasteTable", input).Return()

		err := s.RenderWaste(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("JSONFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatJSON, renderer: mr}
		mr.On("OutputWasteJSON", input).Return(nil)

		err := s.RenderWaste(input)
		assert.NoError(t, err)
		mr.AssertExpectations(t)
	})

	t.Run("CSVFormat", func(t *testing.T) {
		mr := new(renderers.MockRenderer)
		s := &service{format: FormatCSV, renderer: mr}
		mr.On("OutputWasteCSV", input).Return(nil)

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

func TestPrintReportSuccess(t *testing.T) {
	mr := new(renderers.MockRenderer)
	s := &service{renderer: mr}

	mr.On("PrintReportSuccess", "/tmp/report.pdf").Return()

	s.PrintReportSuccess("/tmp/report.pdf")
	mr.AssertExpectations(t)
}

func TestUpdatePrintMethods(t *testing.T) {
	mr := new(renderers.MockRenderer)
	s := &service{renderer: mr}

	t.Run("PrintAlreadyLatest", func(t *testing.T) {
		mr.On("PrintAlreadyLatest", "v1.2.3").Return()
		s.PrintAlreadyLatest("v1.2.3")
		mr.AssertExpectations(t)
	})

	t.Run("PrintRateLimitError", func(t *testing.T) {
		mr.On("PrintRateLimitError").Return()
		s.PrintRateLimitError()
		mr.AssertExpectations(t)
	})

	t.Run("PrintUpdateError", func(t *testing.T) {
		err := assert.AnError
		mr.On("PrintUpdateError", err).Return()
		s.PrintUpdateError(err)
		mr.AssertExpectations(t)
	})

	t.Run("RenderVersion", func(t *testing.T) {
		v := model.VersionInfo{Version: "v1"}
		mr.On("RenderVersion", v).Return()
		s.RenderVersion(v)
		mr.AssertExpectations(t)
	})
}

func TestFormatConstants(t *testing.T) {
	if FormatTable != "table" {
		t.Errorf("FormatTable should be 'table', got %q", FormatTable)
	}

	if FormatJSON != "json" {
		t.Errorf("FormatJSON should be 'json', got %q", FormatJSON)
	}

	if FormatCSV != "csv" {
		t.Errorf("FormatCSV should be 'csv', got %q", FormatCSV)
	}
}
