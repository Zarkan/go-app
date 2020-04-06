package app

// BrowserStorage is the interface that describes a web browser storage.
type BrowserStorage interface {
	// Set sets the value to the given key. The value must be json convertible.
	Set(k string, v interface{}) error

	// Get gets the item associated to the given key and store it in the given
	// value.
	// It returns an error if v is not a pointer.
	Get(k string, v interface{}) error

	// Del deletes the item associated with the given key.
	Del(k string)

	// Clear deletes all items.
	Clear()
}
