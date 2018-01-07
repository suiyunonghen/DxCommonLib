package DxCommonLib

type GDxBaseObject struct {
	fSubChilds []interface{} //子对象
	UseData		interface{}  //用户数据
}

type IDxInheritedObject interface {
	SubChild(idx int) interface{}
	SubChildCount() int
	Destroy()
	SubInit()
	LastedSubChild()interface{}				//最后一个继承者
}


func (obj *GDxBaseObject) SubInit(subObj interface{}) {
	if obj.fSubChilds == nil {
		obj.fSubChilds = make([]interface{}, 3)
	}
	for _, v := range obj.fSubChilds {
		if v == subObj {
			return
		}
	}
	obj.fSubChilds = append(obj.fSubChilds, subObj)
}

func (obj *GDxBaseObject)LastedSubChild() interface{}  {
	if obj.fSubChilds == nil{
		return  nil
	}
	return obj.fSubChilds[len(obj.fSubChilds)-1]
}

func (obj *GDxBaseObject)Free()  {
	//执行Destroy过程
	if i := obj.SubChildCount() - 1; i >= 0{
		obj.SubChild(i).(IDxInheritedObject).Destroy()

	}else{
		obj.Destroy()
	}
}

func (obj *GDxBaseObject)Destroy()  {
	//释放的过程，后面继承的重写此方法则可

}

func (obj *GDxBaseObject) SubChild(idx int) interface{} {
	if obj.fSubChilds == nil {
		return nil
	}
	if idx >= 0 && idx < len(obj.fSubChilds) {
		return obj.fSubChilds[idx]
	}
	return nil
}

func (obj *GDxBaseObject) SubChildCount() int {
	if obj.fSubChilds == nil {
		return 0
	}
	return len(obj.fSubChilds)
}