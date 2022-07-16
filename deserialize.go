package mtmap

import (
	"compress/zlib"
	"encoding/binary"
	"errors"
	"github.com/anon55555/mt"
	"io"
)

var (
	ErrInvalidSerializeVer = errors.New("invalid serialize version")
	ErrInvalidContentWidth = errors.New("invalid content width")
	ErrInvalidParamsWidth  = errors.New("invalid params width")
	ErrInvalidNodeMetaVer  = errors.New("invalid node meta version")
	ErrInvalidNameIdMapVer = errors.New("invalid name id mapping version")
	ErrInvalidStaticObjVer = errors.New("invalid static object version")
)

func Deserialize(r io.Reader, idNameMap map[string]mt.Content) *MapBlk {
	var blk = &MapBlk{}

	var ver uint8
	if err := binary.Read(r, binary.BigEndian, &ver); err != nil {
		panic(err)
	}

	if ver != SerializeVer {
		panic(ErrInvalidSerializeVer)
	}

	if err := binary.Read(r, binary.BigEndian, &blk.Flags); err != nil {
		panic(err)
	}

	if err := binary.Read(r, binary.BigEndian, &blk.LightingComplete); err != nil {
		panic(err)
	}

	var contentWidth uint8
	if err := binary.Read(r, binary.BigEndian, &contentWidth); err != nil {
		panic(err)
	}

	if contentWidth != ContentWidth {
		panic(ErrInvalidContentWidth)
	}

	var paramsWidth uint8
	if err := binary.Read(r, binary.BigEndian, &paramsWidth); err != nil {
		panic(err)
	}

	if paramsWidth != ParamsWidth {
		panic(ErrInvalidParamsWidth)
	}

	{
		r, err := zlib.NewReader(r)
		if err != nil {
			panic(err)
		}

		if err := binary.Read(r, binary.BigEndian, &blk.Param0); err != nil {
			panic(err)
		}

		if _, err := io.Copy(io.Discard, r); err != nil {
			panic(err)
		}

		if err := r.Close(); err != nil {
			panic(err)
		}
	}

	blk.NodeMetas = make(map[uint16]*mt.NodeMeta)
	{
		r, err := zlib.NewReader(r)
		if err != nil {
			panic(err)
		}

		var version uint8
		if err := binary.Read(r, binary.BigEndian, &version); err != nil {
			panic(err)
		}

		if version != 0 {
			if version != NodeMetaVer {
				panic(ErrInvalidNodeMetaVer)
			}

			var count uint16
			if err := binary.Read(r, binary.BigEndian, &count); err != nil {
				panic(err)
			}

			for i := uint16(0); i < count; i++ {
				var pos uint16
				if err := binary.Read(r, binary.BigEndian, &pos); err != nil {
					panic(err)
				}

				var num uint32
				if err := binary.Read(r, binary.BigEndian, &num); err != nil {
					panic(err)
				}

				var data = &mt.NodeMeta{}
				data.Fields = make([]mt.NodeMetaField, 0)
				for j := uint32(0); j < num; j++ {
					var field mt.NodeMetaField

					var lenName uint16
					if err := binary.Read(r, binary.BigEndian, &lenName); err != nil {
						panic(err)
					}

					var name = make([]byte, lenName)
					if err := binary.Read(r, binary.BigEndian, &name); err != nil {
						panic(err)
					}
					field.Name = string(name)

					var lenValue uint32
					if err := binary.Read(r, binary.BigEndian, &lenValue); err != nil {
						panic(err)
					}

					var value = make([]byte, lenValue)
					if err := binary.Read(r, binary.BigEndian, &value); err != nil {
						panic(err)
					}
					field.Value = string(value)

					if err := binary.Read(r, binary.BigEndian, &field.Private); err != nil {
						panic(err)
					}

					data.Fields = append(data.Fields, field)
				}

				if err := data.Inv.Deserialize(r); err != nil {
					panic(err)
				}

				blk.NodeMetas[pos] = data
			}
		}

		if _, err := io.Copy(io.Discard, r); err != nil {
			panic(err)
		}

		if err := r.Close(); err != nil {
			panic(err)
		}
	}

	var staticObjVer uint8
	if err := binary.Read(r, binary.BigEndian, &staticObjVer); err != nil {
		panic(err)
	}

	if staticObjVer != StaticObjVer {
		panic(ErrInvalidStaticObjVer)
	}

	var staticObjCount uint16
	if err := binary.Read(r, binary.BigEndian, &staticObjCount); err != nil {
		panic(err)
	}

	blk.StaticObjs = make([]StaticObj, 0)
	for i := uint16(0); i < staticObjCount; i++ {
		var obj StaticObj

		if err := binary.Read(r, binary.BigEndian, &obj.Type); err != nil {
			panic(err)
		}

		var pos [3]int32
		if err := binary.Read(r, binary.BigEndian, &pos); err != nil {
			panic(err)
		}

		obj.Pos = [3]float32{
			float32(pos[0]) / 1000.0,
			float32(pos[1]) / 1000.0,
			float32(pos[2]) / 1000.0,
		}

		var dataLen uint16
		if err := binary.Read(r, binary.BigEndian, &dataLen); err != nil {
			panic(err)
		}

		var data = make([]byte, dataLen)
		if err := binary.Read(r, binary.BigEndian, &data); err != nil {
			panic(err)
		}

		obj.Data = string(data)

		blk.StaticObjs = append(blk.StaticObjs, obj)
	}

	if err := binary.Read(r, binary.BigEndian, &blk.Timestamp); err != nil {
		panic(err)
	}

	var nameIdMapVer uint8
	if err := binary.Read(r, binary.BigEndian, &nameIdMapVer); err != nil {
		panic(err)
	}

	if nameIdMapVer != NameIdMapVer {
		panic(ErrInvalidNameIdMapVer)
	}

	var nameIdMapCount uint16
	if err := binary.Read(r, binary.BigEndian, &nameIdMapCount); err != nil {
		panic(err)
	}

	var nameIdMap = make(map[mt.Content]string)

	for i := uint16(0); i < nameIdMapCount; i++ {
		var id mt.Content
		if err := binary.Read(r, binary.BigEndian, &id); err != nil {
			panic(err)
		}

		var nameLen uint16
		if err := binary.Read(r, binary.BigEndian, &nameLen); err != nil {
			panic(err)
		}

		var name = make([]byte, nameLen)
		if err := binary.Read(r, binary.BigEndian, &name); err != nil {
			panic(err)
		}

		nameIdMap[id] = string(name)
	}

	for i := 0; i < 4096; i++ {
		id := blk.Param0[i]

		name, ok := nameIdMap[id]
		if !ok {
			name = "unknown"
		}

		switch name {
		case "unknown":
			id = mt.Unknown
		case "air":
			id = mt.Air
		case "ignore":
			id = mt.Ignore
		default:
			id, ok = idNameMap[name]
			if !ok {
				id = mt.Unknown
			}
		}

		blk.Param0[i] = id
	}

	return blk
}
