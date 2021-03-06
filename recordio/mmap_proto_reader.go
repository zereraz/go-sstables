package recordio

import (
	"github.com/gogo/protobuf/proto"
)

type MMapProtoReader struct {
	reader *MMapReader
}

func (r *MMapProtoReader) Open() error {
	return r.reader.Open()
}

func (r *MMapProtoReader) ReadNextAt(record proto.Message, offset uint64) (proto.Message, error) {
	bytes, err := r.reader.ReadNextAt(offset)
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(bytes, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *MMapProtoReader) Close() error {
	return r.reader.Close()
}

func NewMMapProtoReaderWithPath(path string) (*MMapProtoReader, error) {
	r, err := NewMemoryMappedReaderWithPath(path)
	if err != nil {
		return nil, err
	}

	return &MMapProtoReader{reader: r}, nil
}
