package memory

import (
	"testing"
	"time"

	"github.com/georgemac/adagio/pkg/repository"
)

func Test_Run_RepositoryTestHarness(t *testing.T) {
	repo := New()

	repository.TestHarness(t, func(now func() time.Time) repository.Repository {
		repo.now = now
		return repo
	})
}
