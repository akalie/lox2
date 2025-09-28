package inner

type ObjType uint8

const (
	OBJ_STRING   ObjType = iota
	OBJ_FUNCTION ObjType = iota
	OBJ_NATIVE   ObjType = iota
)

var objTypeName = map[ObjType]string{
	OBJ_STRING:   "string",
	OBJ_FUNCTION: "function",
	OBJ_NATIVE:   "native",
}

type Obj interface {
	GetType() ObjType
	GetTypeName() string
}

type ObjString struct {
	ttype  ObjType
	length int
	chars  []byte
	hash   uint32
}

func (o ObjString) GetTypeName() string {
	return objTypeName[o.GetType()]
}

func (o ObjString) GetType() ObjType {
	return o.ttype
}

func NewObjString(chars []byte) ObjString {
	return ObjString{
		ttype:  OBJ_STRING,
		chars:  chars,
		length: len(chars),
		hash:   hashString(chars, len(chars)),
	}
}

func hashString(key []byte, length int) uint32 {
	hash := 2166136261

	for i := 0; i < length; i++ {
		hash ^= int(key[i])
		hash *= 16777619
	}
	return uint32(hash)
}

type ObjFunction struct {
	Obj   *Obj
	Arity int
	Chunk *Chunk
	Name  ObjString
}

func (of ObjFunction) GetType() ObjType {
	return OBJ_FUNCTION
}

func (of ObjFunction) GetTypeName() string {
	return objTypeName[of.GetType()]
}

type NativeFn func(argCount byte, args ...Value) Value

type ObjNative struct {
	Obj   *Obj
	NFunk NativeFn
	Name  string
	Arity int
}

func NewObjNative(nfunk NativeFn, name string, arity int) *ObjNative {
	return &ObjNative{
		NFunk: nfunk,
		Name:  name,
		Arity: arity,
	}
}

func (on ObjNative) GetType() ObjType {
	return OBJ_NATIVE
}

func (on ObjNative) GetTypeName() string {
	return objTypeName[on.GetType()]
}

func NewObjFunction(arity int, name ObjString) *ObjFunction {
	chunk := NewChunk()
	return &ObjFunction{
		Arity: arity,
		Chunk: chunk,
		Name:  name,
	}
}
