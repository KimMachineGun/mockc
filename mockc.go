package mockc

// Implement designates the interfaces to be implemented.
func Implement(...interface{}) {}

// SetFieldNamePrefix sets the prefix of the mock's field names.
func SetFieldNamePrefix(string) {}

// SetFieldNameSuffix sets the suffix of the mock's field names.
func SetFieldNameSuffix(string) {}

// SetDestination sets the destination file where the mock will be generated.
func SetDestination(string) {}

// WithConstructor generates the constructor of mock.
// You can set the underlying implementation by passing it to the constructor.
func WithConstructor() {}

// Deprecated: Please use Implement instead.
// Implements designates the interfaces to be implemented.
func Implements(...interface{}) {}
