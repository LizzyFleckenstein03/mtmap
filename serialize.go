package mtmap

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"github.com/anon55555/mt"
	"io"
)

func Serialize(blk *MapBlk, w io.Writer, idNameMap map[string]mt.Content) error {
	if err := binary.Write(w, binary.BigEndian, &SerializeVer); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, &blk.Flags); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, &blk.LightingComplete); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, &ContentWidth); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, &ParamsWidth); err != nil {
		return err
	}

	{
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)

		if err := binary.Write(zw, binary.BigEndian, &blk.Param0); err != nil {
			return err
		}

		if err := zw.Close(); err != nil {
			return err
		}

		if _, err := buf.WriteTo(w); err != nil {
			return err
		}
	}

	{
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)

		if err := binary.Write(zw, binary.BigEndian, &NodeMetaVer); err != nil {
			return err
		}

		var count = uint16(len(blk.NodeMetas))
		if err := binary.Write(zw, binary.BigEndian, &count); err != nil {
			return err
		}

		for pos, data := range blk.NodeMetas {
			if err := binary.Write(zw, binary.BigEndian, &pos); err != nil {
				return err
			}

			var num = uint32(len(data.Fields))
			if err := binary.Write(zw, binary.BigEndian, &num); err != nil {
				return err
			}

			for _, field := range data.Fields {
				var lenName = uint16(len(field.Name))
				if err := binary.Write(zw, binary.BigEndian, &lenName); err != nil {
					return err
				}

				var name = []byte(field.Name)
				if err := binary.Write(zw, binary.BigEndian, &name); err != nil {
					return err
				}

				var lenValue = uint32(len(field.Value))
				if err := binary.Write(zw, binary.BigEndian, &lenValue); err != nil {
					return err
				}

				var value = []byte(field.Value)
				if err := binary.Write(zw, binary.BigEndian, &value); err != nil {
					return err
				}

				if err := binary.Write(zw, binary.BigEndian, &field.Private); err != nil {
					return err
				}
			}

			if err := data.Inv.Serialize(zw); err != nil {
				return err
			}
		}

		if err := zw.Close(); err != nil {
			return err
		}

		if _, err := buf.WriteTo(w); err != nil {
			return err
		}
	}

	if err := binary.Write(w, binary.BigEndian, &StaticObjVer); err != nil {
		return err
	}

	var staticObjCount = uint16(len(blk.StaticObjs))
	if err := binary.Write(w, binary.BigEndian, &staticObjCount); err != nil {
		return err
	}

	for _, obj := range blk.StaticObjs {
		if err := binary.Write(w, binary.BigEndian, &obj.Type); err != nil {
			return err
		}

		var pos = [3]int32{
			int32(obj.Pos[0] * 1000.0),
			int32(obj.Pos[1] * 1000.0),
			int32(obj.Pos[2] * 1000.0),
		}
		if err := binary.Write(w, binary.BigEndian, &pos); err != nil {
			return err
		}

		var dataLen = uint16(len(obj.Data))
		if err := binary.Write(w, binary.BigEndian, &dataLen); err != nil {
			return err
		}

		var data = []byte(obj.Data)
		if err := binary.Write(w, binary.BigEndian, &data); err != nil {
			return err
		}
	}

	if err := binary.Write(w, binary.BigEndian, &blk.Timestamp); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, &NameIdMapVer); err != nil {
		return err
	}

	var nameIdMapCount = uint16(len(idNameMap))
	if err := binary.Write(w, binary.BigEndian, &nameIdMapCount); err != nil {
		return err
	}

	var exists = make(map[mt.Content]struct{})
	for i := 0; i < 4096; i++ {
		exists[blk.Param0[i]] = struct{}{}
	}

	for name, id := range idNameMap {
		if _, ok := exists[id]; !ok {
			continue
		}

		if err := binary.Write(w, binary.BigEndian, &id); err != nil {
			return err
		}

		var nameLen = uint16(len(name))
		if err := binary.Write(w, binary.BigEndian, &nameLen); err != nil {
			return err
		}

		var name = []byte(name)
		if err := binary.Write(w, binary.BigEndian, &name); err != nil {
			return err
		}
	}

	return nil
}
