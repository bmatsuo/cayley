// DO NOT MODIFY THIS FILE
// This file was autogenerated.

package lmdb

import (
	"testing"

	qstest "github.com/bmatsuo/cayley/graph/graphtest/qstest"
)

func init() {
	// TODO: setup default test runner if necessary.
}

func TestQuadStoreCreate(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreCreate", qstest.TestQuadStoreCreate)
}

func TestQuadStoreLoadFixture(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreLoadFixture", qstest.TestQuadStoreLoadFixture)
}

func TestQuadStoreRemoveQuad(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreRemoveQuad", qstest.TestQuadStoreRemoveQuad)
}

func TestQuadStoreNodesAllIterator(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreNodesAllIterator", qstest.TestQuadStoreNodesAllIterator)
}

func TestQuadStoreQuadsAllIterator(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreQuadsAllIterator", qstest.TestQuadStoreQuadsAllIterator)
}

func TestQuadStoreQuadIterator(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreQuadIterator", qstest.TestQuadStoreQuadIterator)
}

func TestQuadStoreQuadIteratorAnd(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreQuadIteratorAnd", qstest.TestQuadStoreQuadIteratorAnd)
}

func TestQuadStoreQuadIteratorReset(t *testing.T) {
	(Runner).Run(t, "TestQuadStoreQuadIteratorReset", qstest.TestQuadStoreQuadIteratorReset)
}
