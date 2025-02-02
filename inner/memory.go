package inner

func freeObjects(vm *Vm) {
	vm.objects = nil
	return
	for object := vm.objects; object != nil; {
		next := object.next
		freeObject(object)
		object = next
	}
}

func freeObject(object *ObjValue) {
	//todo
	//switch object.v.(type) {
	//case *ObjString:
	//	freeString()
	//}
}
