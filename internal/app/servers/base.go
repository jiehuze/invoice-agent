package servers

type IServer interface {
	Start() error
	Stop() error
}
