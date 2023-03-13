package byteCodec

import (
	"encoding/binary"
	"io"
	"math"
)

// ----- 反序列化 -----

func Uint16Big(p []byte) uint16 {
	return binary.BigEndian.Uint16(p)
}

func Uint24Big(p []byte) uint32 {
	return uint32(p[2]) | uint32(p[1])<<8 | uint32(p[0])<<16
}

func Uint32Big(p []byte) (ret uint32) {
	return binary.BigEndian.Uint32(p)
}

func Uint64Big(p []byte) (ret uint64) {
	return binary.BigEndian.Uint64(p)
}

func Float64Big(p []byte) (ret float64) {
	a := binary.BigEndian.Uint64(p)
	return math.Float64frombits(a)
}

func Uint32Little(p []byte) (ret uint32) {
	return binary.LittleEndian.Uint32(p)
}
func Uint16Little(p []byte) (ret uint16) {
	return binary.LittleEndian.Uint16(p)
}

// ----- 序列化 -----

func PutUint16Big(out []byte, in uint16) {
	binary.BigEndian.PutUint16(out, in)
}

func EncodeUint16Big(in uint16) []byte {
	out := make([]byte, 2)
	binary.BigEndian.PutUint16(out, in)
	return out
}

func PutUint24Big(out []byte, in uint32) {
	out[0] = byte(in >> 16)
	out[1] = byte(in >> 8)
	out[2] = byte(in)
}

func EncodeUint24Big(in uint32) []byte {
	out := make([]byte, 3)
	out[0] = byte(in >> 16)
	out[1] = byte(in >> 8)
	out[2] = byte(in)
	return out
}

func PutUint32Big(out []byte, in uint32) {
	binary.BigEndian.PutUint32(out, in)
}

func EncodeUint32Big(in uint32) []byte {
	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, in)
	return out
}

func PutUint64Big(out []byte, in uint64) {
	binary.BigEndian.PutUint64(out, in)
}

func EncodeUint64Big(in uint64) []byte {
	out := make([]byte, 8)
	binary.BigEndian.PutUint64(out, in)
	return out
}

func PutFloat64Big(out []byte, in float64) {
	binary.BigEndian.PutUint64(out, math.Float64bits(in))
}

func EncodeFloat64Big(in float64) []byte {
	out := make([]byte, 8)
	binary.BigEndian.PutUint64(out, math.Float64bits(in))
	return out
}

func PutUint32Little(out []byte, in uint32) {
	binary.LittleEndian.PutUint32(out, in)
}

func EncodeUint32Little(in uint32) []byte {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint32(out, in)
	return out
}

func PutUint16Little(out []byte, in uint16) {
	binary.LittleEndian.PutUint16(out, in)
}

func EncodeUint16Little(in uint16) []byte {
	out := make([]byte, 2)
	binary.LittleEndian.PutUint16(out, in)
	return out
}

func WriteBig(writer io.Writer, in interface{}) error {
	return binary.Write(writer, binary.BigEndian, in)
}

func WriteLittle(writer io.Writer, in interface{}) error {
	return binary.Write(writer, binary.LittleEndian, in)
}
