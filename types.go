package mtmap

import (
	"github.com/anon55555/mt"
)

type MapBlk struct {
	mt.MapBlk
	Flags            MapBlkFlags
	LightingComplete uint16
	StaticObjs       []StaticObj
	Timestamp        uint32
}

type MapBlkFlags uint8

const (
	IsUnderground MapBlkFlags = 1 << iota
	DayNightDiffers
	NotGenerated = 1 << 4
)

var (
	SerializeVer uint8 = 28
	ContentWidth uint8 = 2
	ParamsWidth  uint8 = 2
	NodeMetaVer  uint8 = 2
	StaticObjVer uint8 = 0
	NameIdMapVer uint8 = 0
)

type StaticObj struct {
	Type uint8
	Pos  [3]float32
	Data string
}
