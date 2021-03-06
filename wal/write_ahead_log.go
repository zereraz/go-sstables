package wal

import (
	"errors"
	"github.com/thomasjungblut/go-sstables/recordio"
)

const DefaultMaxWalSize uint64 = 128 * 1024 * 1024 // 128mb

type WriteAheadLogReplayI interface {
	// Replays the whole WAL from start, calling the given process function
	// for each record in guaranteed order.
	Replay(process func(record []byte) error) error
}

type WriteAheadLogAppendI interface {
	recordio.CloseableI
	// Appends a given record and execute fsync to guarantee the persistence of the record.
	// Has considerably less throughput than Append.
	AppendSync(record []byte) error
}

type WriteAheadLogCleanI interface {
	// Removes all WAL files and the directory it is contained in
	Clean() error
}

type WriteAheadLogCompactI interface {
	// This should compact the WAL, but isn't properly implemented just yet
	Compact() error
}

type WriteAheadLogI interface {
	WriteAheadLogAppendI
	WriteAheadLogReplayI
	WriteAheadLogCleanI
}

type WriteAheadLog struct {
	*Appender
	*Replayer
	*Cleaner
}

func NewWriteAheadLog(opts *Options) (*WriteAheadLog, error) {
	appender, err := NewAppender(opts)
	if err != nil {
		return nil, err
	}
	replayer, err := NewReplayer(opts)
	if err != nil {
		return nil, err
	}
	return &WriteAheadLog{
		Appender: appender,
		Replayer: replayer,
		Cleaner:  NewCleaner(opts),
	}, nil
}

func NewWriteAheadLogOptions(walOptions ...Option) (*Options, error) {
	opts := &Options{
		basePath:       "",
		maxWalFileSize: DefaultMaxWalSize,
		writerFactory: func(path string) (recordio.WriterI, error) {
			return recordio.NewFileWriterWithPath(path)
		},
		readerFactory: func(path string) (recordio.ReaderI, error) {
			return recordio.NewFileReaderWithPath(path)
		},
	}

	for _, walOption := range walOptions {
		walOption(opts)
	}

	if opts.basePath == "" {
		return nil, errors.New("basePath was not supplied")
	}

	return opts, nil
}

// options

type Options struct {
	maxWalFileSize uint64
	basePath       string
	// TODO(thomas): this should be ideally in a writer-only option
	writerFactory func(path string) (recordio.WriterI, error)
	// TODO(thomas): this should be ideally in a reader-only option
	readerFactory func(path string) (recordio.ReaderI, error)
}

type Option func(*Options)

func BasePath(p string) Option {
	return func(args *Options) {
		args.basePath = p
	}
}

func MaximumWalFileSizeBytes(p uint64) Option {
	return func(args *Options) {
		args.maxWalFileSize = p
	}
}

func WriterFactory(writerFactory func(path string) (recordio.WriterI, error)) Option {
	return func(args *Options) {
		args.writerFactory = writerFactory
	}
}

func ReaderFactory(readerFactory func(path string) (recordio.ReaderI, error)) Option {
	return func(args *Options) {
		args.readerFactory = readerFactory
	}
}
