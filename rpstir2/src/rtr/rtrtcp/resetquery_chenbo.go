package rtrtcp

import (
	"fmt"
	"model"
	"os"
	"rtr/db"
	rtrmodel "rtr/model"
	"strconv"
	"time"
	"uint128"

	belogs "github.com/astaxie/beego/logs"
)

type Ipv4RtrPrefixMap struct {
	Asn       uint32            `json:"asn"`
	PrefixMap map[uint32]uint32 `json:"prefixMap"`
}

type Ipv6RtrPrefixMap struct {
	Asn            uint32            `json:"asn"`
	PrefixMapKey   []uint128.Uint128 `json:"prefixMapKey"`
	PrefixMapValue []uint128.Uint128 `json:"prefixMapValue"`
}

func MzrProcessResetQuery(rtrPduModel rtrmodel.RtrPduModel) (resetResponses []rtrmodel.RtrPduModel, err error) {

	start_database_v4 := time.Now().UnixNano()
	rawIpv4RtrFulls, sessionId, serialNumber, err := db.MzrGetIpv4RtrFullRawAndSessionIdAndSerialNumber()
	end_database_v4 := time.Now().UnixNano()
	database_v4 := end_database_v4 - start_database_v4

	start_encode_v4 := time.Now().UnixNano()
	ipv4RtrPrefixMaps, ipv4AsnZeroRawData, err := ConvertIpv4RawDataToPrefixMaps(rawIpv4RtrFulls)
	end_encode_v4 := time.Now().UnixNano()
	encode_time_v4 := end_encode_v4 - start_encode_v4
	fmt.Printf("ipv4 encode time : %v \n", encode_time_v4)
	f1, err := os.OpenFile("/home/cnic/rpki/newtest/chenbo/current/ipv4_encode_time.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f1.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f1.WriteString(fmt.Sprintf("%v \n", encode_time_v4))

	if err != nil {
		belogs.Error("MzrProcessResetQuery(): Get Ipv4RtrFulls fail: ", err)
		return resetResponses, err
	}

	start_database_v6 := time.Now().UnixNano()
	rawIpv6RtrFulls, sessionId, serialNumber, err := db.MzrGetIpv6RtrFullRawAndSessionIdAndSerialNumber()
	end_database_v6 := time.Now().UnixNano()
	database_v6 := end_database_v6 - start_database_v6

	start_encode_v6 := time.Now().UnixNano()
	ipv6RtrPrefixMaps, ipv6AsnZeroRawData, err := ConvertIpv6RawDataToPrefixMaps(rawIpv6RtrFulls)
	end_encode_v6 := time.Now().UnixNano()
	encode_time_v6 := end_encode_v6 - start_encode_v6
	fmt.Printf("ipv6 encode time : %v \n", encode_time_v6)

	fmt.Printf("ipv4 database time : %v\nipv6 database tiem : %v\n", database_v4, database_v6)
	//f2, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_hroa_current_ipv4.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	f2, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_hroa_current_ipv4.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f2.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f2.WriteString(fmt.Sprintf("%v \n", database_v4))

	//f, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_hroa_current_ipv6.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	f, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_hroa_current_ipv6.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f.WriteString(fmt.Sprintf("%v \n", database_v6))
	f3, err := os.OpenFile("/home/cnic/rpki/newtest/chenbo/current/ipv6_encode_time.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f3.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f3.WriteString(fmt.Sprintf("%v \n", encode_time_v6))

	if err != nil {
		belogs.Error("MzrProcessResetQuery(): Open Ipv6RtrFulls fail: ", err)
		return resetResponses, err
	}

	pduModels := mzrConvertPrefixMapsToRtrPduModels(ipv4RtrPrefixMaps, ipv6RtrPrefixMaps, 1)

	pduModels = append(pduModels, ConvertRawDataToPduModels(ipv4AsnZeroRawData, rtrmodel.PROTOCOL_VERSION_1)...)

	pduModels = append(pduModels, ConvertRawDataToPduModels(ipv6AsnZeroRawData, rtrmodel.PROTOCOL_VERSION_1)...)

	rtrPduModels, err := assembleResetResponses(pduModels, rtrmodel.PROTOCOL_VERSION_1, sessionId, serialNumber)
	if err != nil {
		belogs.Error("MzrProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}
	return rtrPduModels, nil
}

func ConvertIpv4RawDataToPrefixMaps(raw []model.MzrLabRpkiRtrRawData) ([]*Ipv4RtrPrefixMap, []model.MzrLabRpkiRtrRawData, error) {
	var (
		rawAsnZero []model.MzrLabRpkiRtrRawData
		rawNormal  []model.MzrLabRpkiRtrRawData
	)

	var v4_zero_count int
	var v4_count int
	for _, data := range raw {
		if data.Asn == 0 {
			rawAsnZero = append(rawAsnZero, data)
			v4_zero_count += 1
		} else {
			rawNormal = append(rawNormal, data)
			v4_count += 1
		}
	}
	//fmt.Printf("ipv4 zero count : %v \n ipv4 count : %v \n", v4_zero_count, v4_count)
	ipv4RtrFulls := mzrConvertRawDataToRtrFulls(rawNormal)
	ipv4RtrPrefixMaps, err := mzrConvertIpv4RtrFullsToPrefixMaps(ipv4RtrFulls)
	if err != nil {
		belogs.Error("MzrProcessResetQuery(): Convert Ipv4RtrFulls to PrefixMap error")
		return nil, nil, err
	}
	return ipv4RtrPrefixMaps, rawAsnZero, nil
}

func ConvertIpv6RawDataToPrefixMaps(raw []model.MzrLabRpkiRtrRawData) ([]*Ipv6RtrPrefixMap, []model.MzrLabRpkiRtrRawData, error) {
	var (
		rawAsnZero []model.MzrLabRpkiRtrRawData
		rawNormal  []model.MzrLabRpkiRtrRawData
	)

	var v6_zero_count int
	var v6_count int
	for _, data := range raw {
		if data.Asn == 0 {
			rawAsnZero = append(rawAsnZero, data)
			v6_zero_count += 1
		} else {
			rawNormal = append(rawNormal, data)
			v6_count += 1
		}
	}
	//fmt.Printf("ipv6 zero count : %v \nipv6 count : %v \n", v6_zero_count, v6_count)
	ipv6RtrFulls := mzrConvertRawDataToRtrFulls(rawNormal)
	ipv6RtrPrefixMaps, err := mzrConvertIpv6RtrFullsToPrefixMaps(ipv6RtrFulls)
	if err != nil {
		belogs.Error("MzrProcessResetQuery(): Convert Ipv6RtrFulls to PrefixMap error")
		return nil, nil, err
	}
	return ipv6RtrPrefixMaps, rawAsnZero, nil
}

func mzrConvertPrefixMapsToRtrPduModels(ipv4PrefixMaps []*Ipv4RtrPrefixMap, ipv6PrefixMaps []*Ipv6RtrPrefixMap, protocolVersion uint8) []rtrmodel.RtrPduModel {
	var rtrPduModels []rtrmodel.RtrPduModel

	rtr_count_v4 := 0

	if len(ipv4PrefixMaps) > 0 {
		ipv4RtrPduModels := mzrConvertIpv4PrefixMapsToRtrPduModel(ipv4PrefixMaps, protocolVersion)
		for _, m := range ipv4RtrPduModels {
			rtrPduModels = append(rtrPduModels, m)
			rtr_count_v4 += 1
		}
	}

	rtr_count_v6 := 0

	if len(ipv6PrefixMaps) > 0 {
		ipv6RtrPduModels := mzrConvertIpv6PrefixMapsToRtrPduModel(ipv6PrefixMaps, protocolVersion)
		for _, m := range ipv6RtrPduModels {
			rtrPduModels = append(rtrPduModels, m)
			rtr_count_v6 += 1
		}
	}

	fmt.Printf("ipv4 rtr count : %v \n", rtr_count_v4)
	fmt.Printf("ipv6 rtr count : %v \n", rtr_count_v6)

	return rtrPduModels
}

func mzrConvertIpv4RtrFullsToPrefixMaps(ipv4RtrFulls []*RtrFull) ([]*Ipv4RtrPrefixMap, error) {
	var ipv4RtrPrefixMaps []*Ipv4RtrPrefixMap
	if len(ipv4RtrFulls) > 0 {
		keyMask := make([]uint64, 33)
		for j := 0; j < 33; j++ {
			keyMask[j] = uint64((1 << 32) - (1 << (32 - j/5*5)))
		}
		for _, ipv4RtrFull := range ipv4RtrFulls {
			prefixMap := make(map[uint32]uint32)
			for _, prefix := range ipv4RtrFull.PrefixList {
				prefixAddr, err := strconv.ParseUint(prefix.Address, 16, 32)
				if err != nil {
					belogs.Error("mzrConvertIpv4RtrFullsToPrefixMaps(): parse prefixAddr fail")
					return nil, err
				}
				prefixLength := prefix.PrefixLength
				prefixKey := (prefixAddr) & keyMask[prefixLength]
				subtreePath := (prefixAddr ^ prefixKey) >> (32 - prefixLength)
				subtreeMask := uint32(1 << ((1 << (prefixLength % 5)) + subtreePath))
				mapKey := uint32((1 << (prefixLength / 5 * 5)) + (prefixKey >> (32 - (prefixLength / 5 * 5))))
				_, ok := prefixMap[mapKey]
				if ok {
					prefixMap[mapKey] |= subtreeMask
				} else {
					prefixMap[mapKey] = subtreeMask
				}
			}
			ipv4RtrPrefixMaps = append(ipv4RtrPrefixMaps, &Ipv4RtrPrefixMap{
				Asn:       uint32(ipv4RtrFull.Asn),
				PrefixMap: prefixMap,
			})
		}
	}
	return ipv4RtrPrefixMaps, nil
}

func mzrConvertIpv4PrefixMapsToRtrPduModel(ipv4RtrPrefixMaps []*Ipv4RtrPrefixMap, protocolVersion uint8) []*rtrmodel.MzrRtrIpv4PrefixModel {
	var allRtrPduModel []*rtrmodel.MzrRtrIpv4PrefixModel
	flag := uint32(1)
	for _, prefixMap := range ipv4RtrPrefixMaps {
		for k, v := range prefixMap.PrefixMap {
			m := rtrmodel.NewMzrRtrIpv4PrefixModel(protocolVersion, k, v|flag, prefixMap.Asn)
			allRtrPduModel = append(allRtrPduModel, m)
		}
	}
	return allRtrPduModel
}

func mzrConvertIpv6RtrFullsToPrefixMaps(ipv6RtrFulls []*RtrFull) ([]*Ipv6RtrPrefixMap, error) {
	var ipv6RtrPrefixMaps []*Ipv6RtrPrefixMap
	if len(ipv6RtrFulls) > 0 {
		keyMask := make([]uint128.Uint128, 129)
		watchKeyMask := make([]string, 129)
		for j := 0; j < 129; j++ {
			keyMask[j] = uint128.NewUint128(0).Not().Rsh(uint(128 - j/5*5)).Lsh(uint(128 - j/5*5))
			watchKeyMask[j] = keyMask[j].String(2)
		}
		for _, ipv6RtrFull := range ipv6RtrFulls {
			var prefixMapKey []uint128.Uint128
			var prefixMapValue []uint128.Uint128
			//var aa []string
			//for _, prefix := range ipv6RtrFull.PrefixList {
			//	aa = append(aa, fmt.Sprintf("%032x %d", prefix.Address, prefix.PrefixLength))
			//}
			//a := fmt.Sprintf("%d %d %s", ipv6RtrFull.Asn, ipv6RtrFull.Length, strings.Join(aa, " "))
			//fmt.Println(a)
			for _, prefix := range ipv6RtrFull.PrefixList {
				prefixAddr := uint128.FromString(prefix.Address)
				prefixLength := prefix.PrefixLength
				prefixKey := prefixAddr.And(keyMask[prefixLength])
				subtreePath := prefixAddr.Xor(prefixKey).Rsh(uint(128 - prefixLength)).Uint64()
				subtreeMask := uint128.NewUint128(1).Lsh(uint((1 << (prefixLength % 5)) + subtreePath))
				mapKey := uint128.NewUint128(1).Lsh(uint(prefixLength / 5 * 5)).Add(prefixKey.Rsh(uint(128 - (prefixLength / 5 * 5))))
				found := false
				for j, k := range prefixMapKey {
					if k.Cmp(mapKey) == 0 {
						prefixMapValue[j] = prefixMapValue[j].Or(subtreeMask)
						found = true
					}
				}
				if !found {
					prefixMapKey = append(prefixMapKey, mapKey)
					prefixMapValue = append(prefixMapValue, subtreeMask)
				}
			}
			ipv6RtrPrefixMaps = append(ipv6RtrPrefixMaps, &Ipv6RtrPrefixMap{
				Asn:            uint32(ipv6RtrFull.Asn),
				PrefixMapKey:   prefixMapKey,
				PrefixMapValue: prefixMapValue,
			})
		}
	}
	return ipv6RtrPrefixMaps, nil
}
func mzrConvertIpv6PrefixMapsToRtrPduModel(ipv6RtrPrefixMaps []*Ipv6RtrPrefixMap, protocolVersion uint8) []*rtrmodel.MzrRtrIpv6PrefixModel {
	var allRtrPduModel []*rtrmodel.MzrRtrIpv6PrefixModel
	for _, prefixMap := range ipv6RtrPrefixMaps {
		for i := 0; i < len(prefixMap.PrefixMapKey); i++ {
			v := prefixMap.PrefixMapValue[i].Uint64()
			m := rtrmodel.NewMzrRtrIpv6PrefixModel(protocolVersion, prefixMap.PrefixMapKey[i], uint32(v|1), prefixMap.Asn)
			allRtrPduModel = append(allRtrPduModel, m)
			//belogs.Debug("%d %d %08x", asn, prefixMapKey[i].String(), uint8(v|1))
		}
	}
	return allRtrPduModel
}

func mzrConvertRawDataToRtrFulls(rawData []model.MzrLabRpkiRtrRawData) []*RtrFull {
	var rtrFulls []*RtrFull
	rtrFullsMap := make(map[uint64][]Prefix)
	for _, data := range rawData {
		rtrFullsMap[data.Asn] = append(rtrFullsMap[data.Asn], Prefix{
			Address:      data.Address,
			PrefixLength: data.PrefixLength,
		})
	}
	for k, v := range rtrFullsMap {
		rtrFull := &RtrFull{
			Asn:        k,
			Length:     uint(len(v)),
			PrefixList: v,
		}
		rtrFulls = append(rtrFulls, rtrFull)
	}
	return rtrFulls
}
