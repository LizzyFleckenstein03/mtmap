package mtmap

import (
	"github.com/anon55555/mt"
)

func (mb MapBlk) PickNode(i int) mt.Node {
	return mt.Node{
		Param0: mb.Param0[i],
		Param1: mb.Param1[i],
		Param2: mb.Param2[i],
	}
}
