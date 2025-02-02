package inner

type ObjType uint8

const (
	OBJ_STRING ObjType = iota
)

var objTypeName = map[ObjType]string{
	OBJ_STRING: "string",
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
