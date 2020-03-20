package mockc

// Implement designates the interfaces to be implemented.
func Implement(...interface{}) {}

// SetFieldNamePrefix sets the prefix of the mock's field names.
func SetFieldNamePrefix(string) {}

// SetFieldNameSuffix sets the suffix of the mock's field names.
func SetFieldNameSuffix(string) {}

// Deprecated: Please use Implement instead.
// Implements designates the interfaces to be implemented.
func Implements(...interface{}) {}
