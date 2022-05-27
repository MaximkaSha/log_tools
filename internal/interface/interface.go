package Interface

type Storager interface {
	InsertData(typeVar string, name string, value string) int
}
