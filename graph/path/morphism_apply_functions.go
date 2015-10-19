// Copyright 2014 The Cayley Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package path

import (
	"github.com/bmatsuo/cayley/graph"
	"github.com/bmatsuo/cayley/graph/iterator"
	"github.com/bmatsuo/cayley/quad"
)

// join puts two iterators together by intersecting their result sets with an AND
// Since we're using an and iterator, it's a good idea to put the smallest result
// set first so that Next() produces fewer values to check Contains().
func join(qs graph.QuadStore, itL, itR graph.Iterator) graph.Iterator {
	and := iterator.NewAnd(qs)
	and.AddSubIterator(itL)
	and.AddSubIterator(itR)

	return and
}

// isMorphism represents all nodes passed in-- if there are none, this function
// acts as a passthrough for the previous iterator.
func isMorphism(nodes ...string) morphism {
	return morphism{
		Name:     "is",
		Reversal: func() morphism { return isMorphism(nodes...) },
		Apply: func(qs graph.QuadStore, in graph.Iterator) graph.Iterator {
			if len(nodes) == 0 {
				// Acting as a passthrough here is equivalent to
				// building a NodesAllIterator to Next() or Contains()
				// from here as in previous versions.
				return in
			}

			isNodes := qs.FixedIterator()
			for _, n := range nodes {
				isNodes.Add(qs.ValueOf(n))
			}

			// Anything with fixedIterators will usually have a much
			// smaller result set, so join isNodes first here.
			return join(qs, isNodes, in)
		},
	}
}

// hasMorphism is the set of nodes that is reachable via either a *Path, a
// single node.(string) or a list of nodes.([]string).
func hasMorphism(via interface{}, nodes ...string) morphism {
	return morphism{
		Name:     "has",
		Reversal: func() morphism { return hasMorphism(via, nodes...) },
		Apply: func(qs graph.QuadStore, in graph.Iterator) graph.Iterator {
			viaIter := buildViaPath(qs, via).
				BuildIterator()
			ends := func() graph.Iterator {
				if len(nodes) == 0 {
					return qs.NodesAllIterator()
				}

				fixed := qs.FixedIterator()
				for _, n := range nodes {
					fixed.Add(qs.ValueOf(n))
				}
				return fixed
			}()

			trail := iterator.NewLinksTo(qs, viaIter, quad.Predicate)
			dest := iterator.NewLinksTo(qs, ends, quad.Object)

			// If we were given nodes, intersecting with them first will
			// be extremely cheap-- otherwise, it will be the most expensive
			// (requiring iteration over all nodes). We have enough info to
			// make this optimization now since intersections are commutative
			if len(nodes) == 0 { // Where dest involves an All iterator.
				route := join(qs, trail, dest)
				has := iterator.NewHasA(qs, route, quad.Subject)
				return join(qs, in, has)
			}

			// This looks backwards. That's OK-- see the note above.
			route := join(qs, dest, trail)
			has := iterator.NewHasA(qs, route, quad.Subject)
			return join(qs, has, in)
		},
	}
}

func tagMorphism(tags ...string) morphism {
	return morphism{
		Name:     "tag",
		Reversal: func() morphism { return tagMorphism(tags...) },
		Apply: func(qs graph.QuadStore, it graph.Iterator) graph.Iterator {
			for _, t := range tags {
				it.Tagger().Add(t)
			}
			return it
		},
		tags: tags,
	}
}

// outMorphism iterates forward one RDF triple or via an entire path.
func outMorphism(via ...interface{}) morphism {
	return morphism{
		Name:     "out",
		Reversal: func() morphism { return inMorphism(via...) },
		Apply: func(qs graph.QuadStore, it graph.Iterator) graph.Iterator {
			path := buildViaPath(qs, via...)
			return inOutIterator(path, it, false)
		},
	}
}

// inMorphism iterates backwards one RDF triple or via an entire path.
func inMorphism(via ...interface{}) morphism {
	return morphism{
		Name:     "in",
		Reversal: func() morphism { return outMorphism(via...) },
		Apply: func(qs graph.QuadStore, it graph.Iterator) graph.Iterator {
			path := buildViaPath(qs, via...)
			return inOutIterator(path, it, true)
		},
	}
}

// iteratorMorphism simply tacks the input iterator onto the chain.
func iteratorMorphism(it graph.Iterator) morphism {
	return morphism{
		Name:     "iterator",
		Reversal: func() morphism { return iteratorMorphism(it) },
		Apply: func(qs graph.QuadStore, subIt graph.Iterator) graph.Iterator {
			return join(qs, it, subIt)
		},
	}
}

// andMorphism sticks a path onto the current iterator chain.
func andMorphism(p *Path) morphism {
	return morphism{
		Name:     "and",
		Reversal: func() morphism { return andMorphism(p) },
		Apply: func(qs graph.QuadStore, itL graph.Iterator) graph.Iterator {
			itR := p.BuildIteratorOn(qs)

			return join(qs, itL, itR)
		},
	}
}

// orMorphism is the union, vice intersection, of a path and the current iterator.
func orMorphism(p *Path) morphism {
	return morphism{
		Name:     "or",
		Reversal: func() morphism { return orMorphism(p) },
		Apply: func(qs graph.QuadStore, itL graph.Iterator) graph.Iterator {
			itR := p.BuildIteratorOn(qs)

			or := iterator.NewOr()
			or.AddSubIterator(itL)
			or.AddSubIterator(itR)
			return or
		},
	}
}

func followMorphism(p *Path) morphism {
	return morphism{
		Name:     "follow",
		Reversal: func() morphism { return followMorphism(p.Reverse()) },
		Apply: func(qs graph.QuadStore, base graph.Iterator) graph.Iterator {
			return p.Morphism()(qs, base)
		},
	}
}

// exceptMorphism removes all results on p.(*Path) from the current iterators.
func exceptMorphism(p *Path) morphism {
	return morphism{
		Name:     "except",
		Reversal: func() morphism { return exceptMorphism(p) },
		Apply: func(qs graph.QuadStore, base graph.Iterator) graph.Iterator {
			in := p.BuildIteratorOn(qs)
			allNodes := qs.NodesAllIterator()
			notIn := iterator.NewNot(in, allNodes)

			return join(qs, base, notIn)
		},
	}
}

func saveMorphism(via interface{}, tag string) morphism {
	return morphism{
		Name:     "save",
		Reversal: func() morphism { return saveMorphism(via, tag) },
		Apply: func(qs graph.QuadStore, it graph.Iterator) graph.Iterator {
			return buildSave(qs, via, tag, it, false)
		},
		tags: []string{tag},
	}
}

func saveReverseMorphism(via interface{}, tag string) morphism {
	return morphism{
		Name:     "saver",
		Reversal: func() morphism { return saveReverseMorphism(via, tag) },
		Apply: func(qs graph.QuadStore, it graph.Iterator) graph.Iterator {
			return buildSave(qs, via, tag, it, true)
		},
		tags: []string{tag},
	}
}

func buildSave(
	qs graph.QuadStore, via interface{},
	tag string, from graph.Iterator, reverse bool,
) graph.Iterator {

	allNodes := qs.NodesAllIterator()
	allNodes.Tagger().Add(tag)

	start, goal := quad.Subject, quad.Object
	if reverse {
		start, goal = goal, start
	}
	viaIter := buildViaPath(qs, via).
		BuildIterator()

	dest := iterator.NewLinksTo(qs, allNodes, goal)
	trail := iterator.NewLinksTo(qs, viaIter, quad.Predicate)

	route := join(qs, trail, dest)
	save := iterator.NewHasA(qs, route, start)

	return join(qs, from, save)
}

func inOutIterator(viaPath *Path, from graph.Iterator, inIterator bool) graph.Iterator {
	start, goal := quad.Subject, quad.Object
	if inIterator {
		start, goal = goal, start
	}

	viaIter := viaPath.BuildIterator()

	source := iterator.NewLinksTo(viaPath.qs, from, start)
	trail := iterator.NewLinksTo(viaPath.qs, viaIter, quad.Predicate)

	route := join(viaPath.qs, source, trail)

	return iterator.NewHasA(viaPath.qs, route, goal)
}

func buildViaPath(qs graph.QuadStore, via ...interface{}) *Path {
	if len(via) == 0 {
		return PathFromIterator(qs, qs.NodesAllIterator())
	} else if len(via) == 1 {
		v := via[0]
		switch p := v.(type) {
		case *Path:
			return p
		case string:
			return StartPath(qs, p)
		default:
			panic("Invalid type passed to buildViaPath.")
		}
	}
	var strings []string
	for _, s := range via {
		if str, ok := s.(string); ok {
			strings = append(strings, str)
		} else {
			panic("Non-string type passed to long Via path")
		}
	}
	return StartPath(qs, strings...)
}
