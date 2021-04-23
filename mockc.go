package mockc

// Implement designates the interfaces to be implemented.
func Implement(i ...interface{}) {}

// SetFieldNamePrefix sets the prefix of the mock's field names.
func SetFieldNamePrefix(prefix string) {}

// SetFieldNameSuffix sets the suffix of the mock's field names.
func SetFieldNameSuffix(suffix string) {}

// SetDestination sets the destination file where the mock will be generated.
// SetDestination only uses the file name of the given destination.
// If the destination is not a go file, the mock generation will fail.
func SetDestination(destination string) {}

// WithConstructor generates the constructor of mock.
// You can set the underlying implementation by passing real implementation to the constructor.
//
// Internally, the WithConstructor is equivalent to the SetConstructorName("New" + MOCK_NAME)
//
// Check below example for details.
// https://github.com/KimMachineGun/mockc/tree/master/examples/with-constructor
func WithConstructor() {}

// SetConstructorName sets the constructor name.
// If the name is empty string, the constructor won't be generated.
func SetConstructorName(name string) {}

// Deprecated: Please use Implement instead.
// Implements designates the interfaces to be implemented.
func Implements(i ...interface{}) {}
