package mtmap

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"github.com/anon55555/mt"
	"io"
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

var SerializeVer uint8 = 28

var (
	ErrInvalidSerializeVer = errors.New("invalid serialize version")
	ErrInvalidContentWidth = errors.New("invalid content width")
	ErrInvalidParamsWidth  = errors.New("invalid params width")
	ErrInvalidNodeMetaVer  = errors.New("invalid node meta version")
	ErrInvalidNameIdMapVer = errors.New("invalid name id mapping version")
	ErrInvalidNode         = errors.New("invalid node")
)

type StaticObj struct {
	Type uint8
	Pos  [3]float32
	Data string
}

func Deserialize(data []byte, idNameMap map[string]mt.Content) (blk *MapBlk, err error) {
	r := bytes.NewReader(data)
	blk = &MapBlk{}

	var ver uint8
	if err := binary.Read(r, binary.BigEndian, &ver); err != nil {
		return nil, err
	}

	if ver != SerializeVer {
		return nil, ErrInvalidSerializeVer
	}

	if err := binary.Read(r, binary.BigEndian, &blk.Flags); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.BigEndian, &blk.LightingComplete); err != nil {
		return nil, err
	}

	var contentWidth uint8
	if err := binary.Read(r, binary.BigEndian, &contentWidth); err != nil {
		return nil, err
	}

	if contentWidth != 2 {
		return nil, ErrInvalidContentWidth
	}

	var paramsWidth uint8
	if err := binary.Read(r, binary.BigEndian, &paramsWidth); err != nil {
		return nil, err
	}

	if paramsWidth != 2 {
		return nil, ErrInvalidParamsWidth
	}

	{
		r, err := zlib.NewReader(r)
		if err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.BigEndian, &blk.Param0); err != nil {
			return nil, err
		}

		if _, err := io.Copy(io.Discard, r); err != nil {
			return nil, err
		}

		if err := r.Close(); err != nil {
			return nil, err
		}
	}

	blk.NodeMetas = make(map[uint16]*mt.NodeMeta)
	{
		r, err := zlib.NewReader(r)
		if err != nil {
			return nil, err
		}

		var version uint8
		if err := binary.Read(r, binary.BigEndian, &version); err != nil {
			return nil, err
		}

		if version != 2 {
			return nil, ErrInvalidNodeMetaVer
		}

		var count uint16
		if err := binary.Read(r, binary.BigEndian, &count); err != nil {
			return nil, err
		}

		for i := uint16(0); i < count; i++ {
			var pos uint16
			if err := binary.Read(r, binary.BigEndian, &pos); err != nil {
				return nil, err
			}

			var num uint32
			if err := binary.Read(r, binary.BigEndian, &num); err != nil {
				return nil, err
			}

			var data = &mt.NodeMeta{}
			data.Fields = make([]mt.NodeMetaField, 0)
			for j := uint32(0); j < num; j++ {
				var field mt.NodeMetaField

				var lenName uint16
				if err := binary.Read(r, binary.BigEndian, &lenName); err != nil {
					return nil, err
				}

				var name = make([]byte, lenName)
				if err := binary.Read(r, binary.BigEndian, &name); err != nil {
					return nil, err
				}
				field.Name = string(name)

				var lenValue uint32
				if err := binary.Read(r, binary.BigEndian, &lenValue); err != nil {
					return nil, err
				}

				var value = make([]byte, lenValue)
				if err := binary.Read(r, binary.BigEndian, &value); err != nil {
					return nil, err
				}
				field.Value = string(value)

				if err := binary.Read(r, binary.BigEndian, &field.Private); err != nil {
					return nil, err
				}

				data.Fields = append(data.Fields, field)
			}

			if err := data.Inv.Deserialize(r); err != nil {
				return nil, err
			}

			blk.NodeMetas[pos] = data
		}

		if _, err := io.Copy(io.Discard, r); err != nil {
			return nil, err
		}

		if err := r.Close(); err != nil {
			return nil, err
		}
	}

	var staticObjVer uint8
	if err := binary.Read(r, binary.BigEndian, &staticObjVer); err != nil {
		return nil, err
	}

	var staticObjCount uint16
	if err := binary.Read(r, binary.BigEndian, &staticObjCount); err != nil {
		return nil, err
	}

	blk.StaticObjs = make([]StaticObj, 0)
	for i := uint16(0); i < staticObjCount; i++ {
		var obj StaticObj

		if err := binary.Read(r, binary.BigEndian, &obj.Type); err != nil {
			return nil, err
		}

		var pos [3]int32
		if err := binary.Read(r, binary.BigEndian, &pos); err != nil {
			return nil, err
		}

		obj.Pos = [3]float32{
			float32(pos[0]) / 1000.0,
			float32(pos[1]) / 1000.0,
			float32(pos[2]) / 1000.0,
		}

		var dataLen uint16
		if err := binary.Read(r, binary.BigEndian, &dataLen); err != nil {
			return nil, err
		}

		var data = make([]byte, dataLen)
		if err := binary.Read(r, binary.BigEndian, &data); err != nil {
			return nil, err
		}

		obj.Data = string(data)

		blk.StaticObjs = append(blk.StaticObjs, obj)
	}

	if err := binary.Read(r, binary.BigEndian, &blk.Timestamp); err != nil {
		return nil, err
	}

	var nameIdMapVer uint8
	if err := binary.Read(r, binary.BigEndian, &nameIdMapVer); err != nil {
		return nil, err
	}

	if nameIdMapVer != 0 {
		return nil, ErrInvalidNameIdMapVer
	}

	var nameIdMapCount uint16
	if err := binary.Read(r, binary.BigEndian, &nameIdMapCount); err != nil {
		return nil, err
	}

	nameIdMap := make(map[mt.Content]string)

	for i := uint16(0); i < nameIdMapCount; i++ {
		var id uint16
		if err := binary.Read(r, binary.BigEndian, &id); err != nil {
			return nil, err
		}

		var nameLen mt.Content
		if err := binary.Read(r, binary.BigEndian, &nameLen); err != nil {
			return nil, err
		}

		var name = make([]byte, nameLen)
		if err := binary.Read(r, binary.BigEndian, &name); err != nil {
			return nil, err
		}

		nameIdMap[mt.Content(id)] = string(name)
	}

	for i := 0; i < 4096; i++ {
		name, ok := nameIdMap[blk.Param0[i]]
		if !ok {
			return nil, ErrInvalidNode
		}

		id, ok := idNameMap[name]
		if !ok {
			return nil, ErrInvalidNode
		}

		blk.Param0[i] = id
	}

	return
}
