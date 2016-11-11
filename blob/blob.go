package blob


type PageInfo struct{
    Title string
    Path string
    Date  string
    Head  string
}

func loadDir(fol string)[]PageInfo){
    //TODO add all the actual loading
    return []PageInfo{}
}


func BlobGetter() func(string, string) map[string]string {
    bb = map[string][]

    
	res = func(fol, file string) map[string]string {
        inner,ok = bb[fol]
        if ! ok {
            inner = LoadDir(fol)
            bb[fol] = inner
        }

        
         
    
        
	}

}
