package cryptlib

import (
	"crypto/des"
	"encoding/base64"
	"errors"
)

func EntryptDesECB(data, key []byte) (string,error) {
	if len(key) > 8 {
		key = key[:8]
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return "",err
	}
	bs := block.BlockSize()
	data = PKCS5Padding(data, bs)
	if len(data)%bs != 0 {
		return "",errors.New("EntryptDesECB Need a multiple of the blocksize")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	return base64.StdEncoding.EncodeToString(out),nil
}

func EntryptDesECBByte(data, key []byte) ([]byte,error) {
	if len(key) > 8 {
		key = key[:8]
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil,err
	}
	bs := block.BlockSize()
	data = PKCS5Padding(data, bs)
	if len(data)%bs != 0 {
		return nil,errors.New("EntryptDesECB Need a multiple of the blocksize")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(out)))
	base64.StdEncoding.Encode(buf, out)
	return buf,nil
}

func DecryptDESECB(d string, key []byte) (string,error) {
	data, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		return "",err
	}
	if len(key) > 8 {
		key = key[:8]
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return "",err
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return "",errors.New("DecryptDES crypto/cipher: input not full blocks")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = PKCS5UnPadding(out)
	return string(out),nil
}

func DecryptDESECBByte(d string, key []byte) ([]byte,error) {
	data, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		return nil,err
	}
	if len(key) > 8 {
		key = key[:8]
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil,err
	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		return nil,errors.New("DecryptDES crypto/cipher: input not full blocks")
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	return PKCS5UnPadding(out),nil
}