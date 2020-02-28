package cryptlib

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"github.com/suiyunonghen/DxCommonLib"
)

type  AES struct {
	fInitVector	[]byte	//CBC初始化向量
}

func NewAes(initvector []byte)*AES  {
	return &AES{fInitVector:initvector}
}

func (aestool *AES)Encrypt(key,value []byte)([]byte,error)  {
	var realkey []byte
	btlen := len(key)
	if btlen < 16{
		realkey = append(key,make([]byte,16-btlen)...)
	}else{
		realkey = key[:16]
	}
	block, err := aes.NewCipher(realkey[:16])
	if err != nil {
		return nil,err
	}
	blockSize := block.BlockSize()
	value = PKCS5Padding(value, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, aestool.fInitVector[:blockSize])
	crypted := make([]byte, len(value))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	blockMode.CryptBlocks(crypted, value)
	return crypted,nil
}

func (aestool *AES)EncryptBase64(key,value []byte)(string,error)  {
	crypted,err := aestool.Encrypt(key,value)
	if err != nil{
		return "",err
	}
	return  base64.StdEncoding.EncodeToString(crypted),nil
}

func (aestool *AES)EncryptHex(key,value []byte)(string,error)  {
	crypted,err := aestool.Encrypt(key,value)
	if err != nil{
		return "",err
	}
	return  DxCommonLib.Bin2Hex(crypted),nil
}

func (aestool *AES)AESDecryptBase64(value,key string) string {
	var realkey []byte
	bt := []byte(key)
	btlen := len(bt)
	if btlen < 16{
		realkey = append(bt,make([]byte,16-btlen)...)
	}else{
		realkey = bt[:16]
	}
	block, err := aes.NewCipher(realkey[:16])
	if err != nil {
		return value
	}
	var vbyte []byte
	vbyte,err = base64.StdEncoding.DecodeString(value)
	if err != nil{
		return value
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block,aestool.fInitVector[:blockSize])
	origData := make([]byte, len(vbyte))
	blockMode.CryptBlocks(origData, vbyte)
	origData = PKCS5UnPadding(origData)
	if origData == nil{
		return value
	}
	return string(origData)
}

func (aestool *AES)AESDecryptWithHex(value,key string) string {
	var realkey []byte
	bt := []byte(key)
	btlen := len(bt)
	if btlen < 16{
		realkey = append(bt,make([]byte,16-btlen)...)
	}else{
		realkey = bt[:16]
	}
	block, err := aes.NewCipher(realkey[:16])
	if err != nil {
		return value
	}
	vbyte := DxCommonLib.Hex2Binary(value)
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block,aestool.fInitVector[:blockSize])
	origData := make([]byte, len(vbyte))
	blockMode.CryptBlocks(origData, vbyte)
	origData = PKCS5UnPadding(origData)
	if origData == nil{
		return value
	}
	return string(origData)
}