package glhf

// AttrFormat defines names and types of OpenGL attributes (vertex format, uniform format, etc.).
//
// Example:
//   AttrFormat{{"position", Vec2}, {"color", Vec4}, {"texCoord": Vec2}}
type AttrFormat []Attr

// Size returns the total size of all attributes of the AttrFormat.
func (af AttrFormat) Size() int {
	total := 0
	for _, attr := range af {
		total += attr.Type.Size()
	}
	return total
}

// Attr represents an arbitrary OpenGL attribute, such as a vertex attribute or a shader
// uniform attribute.
type Attr struct {
	Name string
	Type AttrType
}

// AttrType represents the type of an OpenGL attribute.
type AttrType int

// List of all possible attribute types.
const (
	Int AttrType = iota
	Float
	Vec2
	Vec3
	Vec4
	Mat2
	Mat23
	Mat24
	Mat3
	Mat32
	Mat34
	Mat4
	Mat42
	Mat43
)

// Size returns the size of a type in bytes.
func (at AttrType) Size() int {
	switch at {
	case Int:
		return 4
	case Float:
		return 4
	case Vec2:
		return 2 * 4
	case Vec3:
		return 3 * 4
	case Vec4:
		return 4 * 4
	case Mat2:
		return 2 * 2 * 4
	case Mat23:
		return 2 * 3 * 4
	case Mat24:
		return 2 * 4 * 4
	case Mat3:
		return 3 * 3 * 4
	case Mat32:
		return 3 * 2 * 4
	case Mat34:
		return 3 * 4 * 4
	case Mat4:
		return 4 * 4 * 4
	case Mat42:
		return 4 * 2 * 4
	case Mat43:
		return 4 * 3 * 4
	default:
		panic("size of vertex attribute type: invalid type")
	}
}
