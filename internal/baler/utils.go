package baler

// logger
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type NoopLogger struct{}

func (l *NoopLogger) Info(msg string)  {}
func (l *NoopLogger) Warn(msg string)  {}
func (l *NoopLogger) Error(msg string) {}

// baler config
type OperationType string

const (
	OperationConvert   OperationType = "convert"
	OperationUnconvert OperationType = "unconvert"
)

// TODO: with Logger it should be refactored to an App
// with config, logger
type BalerConfig struct {
	MaxInputFileLines uint64
	MaxInputFileSize  uint64
	MaxOutputFileSize uint64
	MaxBufferSize     uint64
	ExclusionPatterns *[]string
	Operation         OperationType
	FileDelimiter     string
	Verbose           bool
	// baler app attribute(s)
	// TODO: move
	Logger Logger
}
