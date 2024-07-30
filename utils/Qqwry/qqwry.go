package Qqwry

import (
	"EFunc/utils"
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

//go:embed qqwry.dat
var 中国Ip查询库 []byte

// 自动加载数据库文件
func init() {
	if utils.W文件_是否存在("qqwry.dat") {
		err := LoadFile("qqwry.dat")
		if err != nil {
			LoadData(中国Ip查询库)
			fmt.Printf("本地纯真IP数据库,已加载到内存失败已重新加载内置库\n")
		} else {
			fmt.Printf("本地纯真IP数据库,已加载到内存\n")
		}

	} else {
		LoadData(中国Ip查询库)
		fmt.Printf("纯真IP数据库,已加载到内存\n")
	}

}

var (
	data    []byte
	dataLen uint32
	ipCache = &sync.Map{}
)

const (
	indexLen      = 7
	redirectMode1 = 0x01
	redirectMode2 = 0x02
)

type cache struct {
	City string
	Isp  string
}

func byte3ToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}

func gb18030Decode(src []byte) string {
	in := bytes.NewReader(src)
	out := transform.NewReader(in, simplifiedchinese.GB18030.NewDecoder())
	d, _ := io.ReadAll(out)
	return string(d)
}

func Ip查信息(要查询的IP string) (省市 string, 运行商 string, err error) {
	return QueryIP(要查询的IP)
}
func Ip查信息2(要查询的IP string) string {
	var 省份 = ""
	var 运营商 = ""
	省份, 运营商, _ = QueryIP(要查询的IP)
	return 省份 + " " + 运营商
}

// QueryIP 从内存或缓存查询IP
func QueryIP(queryIp string) (city string, isp string, err error) {
	if v, ok := ipCache.Load(queryIp); ok {
		city = v.(cache).City
		isp = v.(cache).Isp
		return
	}
	ip := net.ParseIP(queryIp).To4()
	if ip == nil {
		err = errors.New("ip is not ipv4")
		return
	}
	ip32 := binary.BigEndian.Uint32(ip)
	posA := binary.LittleEndian.Uint32(data[:4])
	posZ := binary.LittleEndian.Uint32(data[4:8])
	var offset uint32 = 0
	for {
		mid := posA + (((posZ-posA)/indexLen)>>1)*indexLen
		buf := data[mid : mid+indexLen]
		_ip := binary.LittleEndian.Uint32(buf[:4])
		if posZ-posA == indexLen {
			offset = byte3ToUInt32(buf[4:])
			buf = data[mid+indexLen : mid+indexLen+indexLen]
			if ip32 < binary.LittleEndian.Uint32(buf[:4]) {
				break
			} else {
				offset = 0
				break
			}
		}
		if _ip > ip32 {
			posZ = mid
		} else if _ip < ip32 {
			posA = mid
		} else if _ip == ip32 {
			offset = byte3ToUInt32(buf[4:])
			break
		}
	}
	if offset <= 0 {
		err = errors.New("ip not found")
		return
	}
	posM := offset + 4
	mode := data[posM]
	var ispPos uint32
	switch mode {
	case redirectMode1:
		posC := byte3ToUInt32(data[posM+1 : posM+4])
		mode = data[posC]
		posCA := posC
		if mode == redirectMode2 {
			posCA = byte3ToUInt32(data[posC+1 : posC+4])
			posC += 4
		}
		for i := posCA; i < dataLen; i++ {
			if data[i] == 0 {
				city = string(data[posCA:i])
				break
			}
		}
		if mode != redirectMode2 {
			posC += uint32(len(city) + 1)
		}
		ispPos = posC
	case redirectMode2:
		posCA := byte3ToUInt32(data[posM+1 : posM+4])
		for i := posCA; i < dataLen; i++ {
			if data[i] == 0 {
				city = string(data[posCA:i])
				break
			}
		}
		ispPos = offset + 8
	default:
		posCA := offset + 4
		for i := posCA; i < dataLen; i++ {
			if data[i] == 0 {
				city = string(data[posCA:i])
				break
			}
		}
		ispPos = offset + uint32(5+len(city))
	}
	if city != "" {
		city = strings.TrimSpace(gb18030Decode([]byte(city)))
	}
	ispMode := data[ispPos]
	if ispMode == redirectMode1 || ispMode == redirectMode2 {
		ispPos = byte3ToUInt32(data[ispPos+1 : ispPos+4])
	}
	if ispPos > 0 {
		for i := ispPos; i < dataLen; i++ {
			if data[i] == 0 {
				isp = string(data[ispPos:i])
				if isp != "" {
					if strings.Contains(isp, "CZ88.NET") {
						isp = ""
					} else {
						isp = strings.TrimSpace(gb18030Decode([]byte(isp)))
					}
				}
				break
			}
		}
	}
	ipCache.Store(queryIp, cache{City: city, Isp: isp})
	return
}

// LoadData 从内存加载IP数据库
func LoadData(database []byte) {
	data = database
	dataLen = uint32(len(data))
}

// LoadFile 从文件加载IP数据库
func LoadFile(filepath string) (err error) {
	data, err = os.ReadFile(filepath)
	if err != nil {
		return
	}
	dataLen = uint32(len(data))
	return
}
