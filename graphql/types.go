package graphql

import (
	"context"
	"fmt"
)

// Type represents a GraphQL type, and should be either an Object, a Scalar,
// or a List
type Type interface {
	String() string

	// isType() is a no-op used to tag the known values of Type, to prevent
	// arbitrary interface{} from implementing Type
	isType()
}

// Scalar is a leaf value.  A custom "Unwrapper" can be attached to the scalar
// so it can have a custom unwrapping (if nil we will use the default unwrapper).
type Scalar struct {
	Type      string
	Unwrapper func(interface{}) (interface{}, error)
}

func (s *Scalar) isType() {}

func (s *Scalar) String() string {
	return s.Type
}

// Enum is a leaf value
type Enum struct {
	Type       string
	Values     []string
	ReverseMap map[interface{}]string
}

func (e *Enum) isType() {}

func (e *Enum) String() string {
	return e.Type
}

func (e *Enum) enumValues() []string {
	return e.Values
}

// Object is a value with several fields
type Object struct {
	Name        string
	Description string
	KeyField    *Field
	Fields      map[string]*Field
	Interfaces  map[string]*Interface //For introspection only
}

func (o *Object) isType() {}

func (o *Object) String() string {
	return o.Name
}

// List is a collection of other values
type List struct {
	Type Type
}

func (l *List) isType() {}

func (l *List) String() string {
	return fmt.Sprintf("[%s]", l.Type)
}

//InputObject defines the object in argument of a query, mutation or subscription
type InputObject struct {
	Name        string
	InputFields map[string]Type
}

func (io *InputObject) isType() {}

func (io *InputObject) String() string {
	return io.Name
}

// NonNull is a non-nullable other value
type NonNull struct {
	Type Type
}

func (n *NonNull) isType() {}

func (n *NonNull) String() string {
	return fmt.Sprintf("%s!", n.Type)
}

// Union is a option between multiple types
type Union struct {
	Name        string
	Description string
	Types       map[string]*Object
}

func (*Union) isType() {}

func (u *Union) String() string {
	return u.Name
}

// Interface defines the graphql interface
type Interface struct {
	Name        string
	Description string
	Types       map[string]*Object
	Fields      map[string]*Field
}

func (*Interface) isType() {}

func (i *Interface) String() string {
	return i.Name
}

// Verify *Scalar, *Object, *List, *InputObject, and *NonNull implement Type
var _ Type = &Scalar{}
var _ Type = &Object{}
var _ Type = &List{}
var _ Type = &InputObject{}
var _ Type = &NonNull{}
var _ Type = &Enum{}
var _ Type = &Union{}
var _ Type = &Interface{}

// A Resolver calculates the value of a field of an object
type Resolver func(ctx context.Context, source, args interface{}, selectionSet *SelectionSet) (interface{}, error)

// A BatchResolver calculates the value of a field for a slice of objects.
type BatchResolver func(ctx context.Context, sources []interface{}, args interface{}, selectionSet *SelectionSet) ([]interface{}, error)

// Field knows how to compute field values of an Object
//
// Fields are responsible for computing their value themselves.
type Field struct {
	Resolve        Resolver
	Type           Type
	Args           map[string]Type
	ParseArguments func(json interface{}) (interface{}, error)

	External  bool
	Expensive bool
}

//Schema used to validate and resolve the queries
type Schema struct {
	Query        Type
	Mutation     Type
	Subscription Type
}

// SelectionSet represents a core GraphQL query
//
// A SelectionSet can contain multiple fields and multiple fragments. For
// example, the query
//
//     {
//       name
//       ... UserFragment
//       memberships {
//         organization { name }
//       }
//     }
//
// results in a root SelectionSet with two selections (name and memberships),
// and one fragment (UserFragment). The subselection `organization { name }`
// is stored in the memberships selection.
//
// Because GraphQL allows multiple fragments with the same name or alias,
// selections are stored in an array instead of a map.
type SelectionSet struct {
	Selections []*Selection
	Fragments  []*FragmentSpread
}

//Selection : A selection represents a part of a GraphQL query
//
// The selection
//
//     me: user(id: 166) { name }
//
// has name "user" (representing the source field to be queried), alias "me"
// (representing the name to be used in the output), args id: 166 (representing
// arguments passed to the source field to be queried), and subselection name
// representing the information to be queried from the resulting object.
type Selection struct {
	Name         string
	Alias        string
	Args         interface{}
	SelectionSet *SelectionSet
	Directives   []*Directive

	UseBatch bool

	// The parsed flag is used to make sure the args for this Selection are only
	// parsed once.
	parsed bool
}

// A FragmentDefinition represents a reusable part of a GraphQL query
//
// The On part of a FragmentDefinition represents the type of source object for which
// this FragmentDefinition should be used. That is not currently implemented in this
// package.
type FragmentDefinition struct {
	Name         string
	On           string
	SelectionSet *SelectionSet
}

// FragmentSpread represents a usage of a FragmentDefinition. Alongside the information
// about the fragment, it includes any directives used at that spread location.
type FragmentSpread struct {
	Fragment   *FragmentDefinition
	Directives []*Directive
}

type Directive struct {
	Name string
	Args interface{}
}
