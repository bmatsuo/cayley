/*
Package qstest provides a test framework and automatic test generation for
graph.QuadStore implementions.  The package provides a QuadStoreImpl type that
an implementation provides.
*/
package qstest

import (
	"fmt"

	// TODO(bmatsuo):
	// Change the testrunner import path if this is going to get mainlined into
	// github.com/google/cayley.
	"github.com/bmatsuo/cayley/graph/graphtest/testrunner"

	"github.com/bmatsuo/cayley/graph"
	"golang.org/x/net/context"
)

type qsVarsKey struct{}

// ContextQuadStore returns the qstest QuadStore stored in ctx or nil if there
// is no QuadStore.
func ContextQuadStore(ctx context.Context) graph.QuadStore {
	vars := testrunner.ContextVars(ctx)
	if vars == nil {
		return nil
	}
	return getQuadStore(vars)
}

func getQuadStore(vs testrunner.Vars) graph.QuadStore {
	qs, _ := vs[qsVarsKey{}].(graph.QuadStore)
	return qs
}

func setQuadStore(vs testrunner.Vars, qs graph.QuadStore) {
	vs[qsVarsKey{}] = qs
}

// QuadStoreImpl defines operations that allow generated tests to run against a
// graph.QuadStore implementation..
type QuadStoreImpl struct {
	Name    string
	NewArgs func(ctx context.Context, name string) (path string, opt graph.Options, err error)
	Close   func(ctx context.Context, name string)
}

var _ testrunner.Stage = &QuadStoreImpl{}

// Setup call NewArgs and creates a graph.QuadStore from the result using the call
//		graph.NewQuadStore(impl.Name, path, opt)
func (impl *QuadStoreImpl) Setup(ctx context.Context, name string) (err error) {
	path, opt, err := impl.NewArgs(ctx, name)
	if err != nil {
		return fmt.Errorf("initializing %s: %v", name, err)
	}

	qs, err := graph.NewQuadStore(impl.Name, path, opt)
	if err != nil {
		impl.Teardown(ctx, name)
		return err
	}

	vars := testrunner.ContextVars(ctx)
	setQuadStore(vars, qs)

	return nil
}

// Teardown calls impl.Close.
func (impl *QuadStoreImpl) Teardown(ctx context.Context, name string) {
	impl.Close(ctx, name)
}
