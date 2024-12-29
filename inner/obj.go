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
	}
}
