package DxCommonLib

//文件格式
type FileCodeMode  uint8


const(
	File_Code_Unknown FileCodeMode = iota
	File_Code_Utf8
	File_Code_Utf16BE
	File_Code_Utf16LE
	File_Code_GBK
)
