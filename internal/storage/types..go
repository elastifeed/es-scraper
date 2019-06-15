package storage

// Storager provides an interface for storing abitrary byte slices
type Storager interface {
	Upload([]byte, string) (string, error)
}
