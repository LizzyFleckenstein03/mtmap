package mtmap

import (
	"github.com/anon55555/mt"
)

func (mb *MapBlk) PeekNode(i int) mt.Node {
	return mt.Node{
		Param0: mb.Param0[i],
		Param1: mb.Param1[i],
		Param2: mb.Param2[i],
	}
}

func (mb *MapBlk) PokeNode(i int, node mt.Node) {
	mb.Param0[i] = node.Param0
	mb.Param1[i] = node.Param1
	mb.Param2[i] = node.Param2
}
