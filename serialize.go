package mtmap

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"github.com/anon55555/mt"
	"io"
)

func Serialize(blk *MapBlk, w io.Writer, nameIdMap map[mt.Content]string) {
	if err := binary.Write(w, binary.BigEndian, &SerializeVer); err != nil {
		panic(err)
	}

	if err := binary.Write(w, binary.BigEndian, &blk.Flags); err != nil {
		panic(err)
	}

	if err := binary.Write(w, binary.BigEndian, &blk.LightingComplete); err != nil {
		panic(err)
	}

	if err := binary.Write(w, binary.BigEndian, &ContentWidth); err != nil {
		panic(err)
	}

	if err := binary.Write(w, binary.BigEndian, &ParamsWidth); err != nil {
		panic(err)
	}

	{
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)

		if err := binary.Write(zw, binary.BigEndian, &blk.Param0); err != nil {
			panic(err)
		}

		if err := zw.Close(); err != nil {
			panic(err)
		}

		if _, err := buf.WriteTo(w); err != nil {
			panic(err)
		}
	}

	{
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)

		var version = NodeMetaVer
		if len(blk.NodeMetas) == 0 {
			version = 0
		}

		if err := binary.Write(zw, binary.BigEndian, &version); err != nil {
			panic(err)
		}

		if version != 0 {
			var count = uint16(len(blk.NodeMetas))
			if err := binary.Write(zw, binary.BigEndian, &count); err != nil {
				panic(err)
			}

			for pos, data := range blk.NodeMetas {
				if err := binary.Write(zw, binary.BigEndian, &pos); err != nil {
					panic(err)
				}

				var num = uint32(len(data.Fields))
				if err := binary.Write(zw, binary.BigEndian, &num); err != nil {
					panic(err)
				}

				for _, field := range data.Fields {
					var lenName = uint16(len(field.Name))
					if err := binary.Write(zw, binary.BigEndian, &lenName); err != nil {
						panic(err)
					}

					var name = []byte(field.Name)
					if err := binary.Write(zw, binary.BigEndian, &name); err != nil {
						panic(err)
					}

					var lenValue = uint32(len(field.Value))
					if err := binary.Write(zw, binary.BigEndian, &lenValue); err != nil {
						panic(err)
					}

					var value = []byte(field.Value)
					if err := binary.Write(zw, binary.BigEndian, &value); err != nil {
						panic(err)
					}

					if err := binary.Write(zw, binary.BigEndian, &field.Private); err != nil {
						panic(err)
					}
				}

				if err := data.Inv.Serialize(zw); err != nil {
					panic(err)
				}
			}
		}

		if err := zw.Close(); err != nil {
			panic(err)
		}

		if _, err := buf.WriteTo(w); err != nil {
			panic(err)
		}
	}

	if err := binary.Write(w, binary.BigEndian, &StaticObjVer); err != nil {
		panic(err)
	}

	var staticObjCount = uint16(len(blk.StaticObjs))
	if err := binary.Write(w, binary.BigEndian, &staticObjCount); err != nil {
		panic(err)
	}

	for _, obj := range blk.StaticObjs {
		if err := binary.Write(w, binary.BigEndian, &obj.Type); err != nil {
			panic(err)
		}

		var pos = [3]int32{
			int32(obj.Pos[0] * 1000.0),
			int32(obj.Pos[1] * 1000.0),
			int32(obj.Pos[2] * 1000.0),
		}
		if err := binary.Write(w, binary.BigEndian, &pos); err != nil {
			panic(err)
		}

		var dataLen = uint16(len(obj.Data))
		if err := binary.Write(w, binary.BigEndian, &dataLen); err != nil {
			panic(err)
		}

		var data = []byte(obj.Data)
		if err := binary.Write(w, binary.BigEndian, &data); err != nil {
			panic(err)
		}
	}

	if err := binary.Write(w, binary.BigEndian, &blk.Timestamp); err != nil {
		panic(err)
	}

	if err := binary.Write(w, binary.BigEndian, &NameIdMapVer); err != nil {
		panic(err)
	}

	var localNameIdMap = make(map[mt.Content]string)

	for i := 0; i < 4096; i++ {
		id := blk.Param0[i]
		if _, ok := localNameIdMap[id]; ok {
			continue
		}

		var name string
		var ok bool

		switch id {
		case mt.Unknown:
			name = "unknown"
		case mt.Air:
			name = "air"
		case mt.Ignore:
			name = "ignore"
		default:
			name, ok = nameIdMap[id]
			if !ok {
				panic(ErrInvalidNodeId{id})
			}
		}

		localNameIdMap[id] = name
	}

	var nameIdMapCount = uint16(len(localNameIdMap))
	if err := binary.Write(w, binary.BigEndian, &nameIdMapCount); err != nil {
		panic(err)
	}

	for id, name := range localNameIdMap {
		if err := binary.Write(w, binary.BigEndian, &id); err != nil {
			panic(err)
		}

		var nameLen = uint16(len(name))
		if err := binary.Write(w, binary.BigEndian, &nameLen); err != nil {
			panic(err)
		}

		var name = []byte(name)
		if err := binary.Write(w, binary.BigEndian, &name); err != nil {
			panic(err)
		}
	}
}
