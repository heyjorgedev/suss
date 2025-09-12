package qrcode

import qrcode "github.com/skip2/go-qrcode"

func CreatePng(text string) ([]byte, error) {
	return qrcode.Encode(text, qrcode.Medium, 256)
}
