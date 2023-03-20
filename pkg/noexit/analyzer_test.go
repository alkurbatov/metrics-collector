package noexit_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/pkg/noexit"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), noexit.Analyzer, "./...")
}
