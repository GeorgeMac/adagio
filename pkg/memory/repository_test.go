package memory

import (
	"testing"

	"github.com/georgemac/adagio/pkg/repository"
)

func Test_Run_RepositoryTestHarness(t *testing.T) {
	repo := New()

	repository.TestHarness(t, repo)
}
