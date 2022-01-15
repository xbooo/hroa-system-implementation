package rtrtcp

import (
	"bytes"
	"encoding/binary"
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

func CompressProcessResetQuery(rtrPduModel rtrmodel.RtrPduModel) (resetResponses []rtrmodel.RtrPduModel, err error) {
	start_database_v4 := time.Now().UnixNano()
	rawIpv4RtrFulls, sessionId, serialNumber, err := db.MzrGetIpv4RtrFullRawAndSessionIdAndSerialNumber()
	end_database_v4 := time.Now().UnixNano()
	database_v4 := end_database_v4 - start_database_v4

	ipv4Tries, ipv4AsnZeroRawData, _ := ConvertIpv4RawDataToTries(rawIpv4RtrFulls)
	if err != nil {
		belogs.Error("MzrProcessResetQuery(): Get Ipv4RtrFulls fail: ", err)
		return resetResponses, err
	}

	start_database_v6 := time.Now().UnixNano()
	rawIpv6RtrFulls, sessionId, serialNumber, err := db.MzrGetIpv6RtrFullRawAndSessionIdAndSerialNumber()
	end_database_v6 := time.Now().UnixNano()
	database_v6 := end_database_v6 - start_database_v6
	fmt.Printf("ipv4 database time : %v\nipv6 database time : %v\n", database_v4, database_v6)
	f1, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_mroa_current_ipv4.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f1.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f1.WriteString(fmt.Sprintf("%v \n", database_v4))

	f, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_mroa_current_ipv6.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f.WriteString(fmt.Sprintf("%v \n", database_v6))

	ipv6Tries, ipv6AsnZeroRawData, _ := ConvertIpv6RawDataToTries(rawIpv6RtrFulls)
	if err != nil {
		belogs.Error("MzrProcessResetQuery(): Open Ipv6RtrFulls fail: ", err)
		return resetResponses, err
	}

	pduModels := ConvertTriesToPduModels(ipv4Tries, ipv6Tries)

	pduModels = append(pduModels, ConvertRawDataToPduModels(ipv4AsnZeroRawData, rtrmodel.PROTOCOL_VERSION_1)...)

	pduModels = append(pduModels, ConvertRawDataToPduModels(ipv6AsnZeroRawData, rtrmodel.PROTOCOL_VERSION_1)...)

	rtrPduModels, err := assembleResetResponses(pduModels, rtrmodel.PROTOCOL_VERSION_1, sessionId, serialNumber)
	if err != nil {
		belogs.Error("MzrProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}
	return rtrPduModels, nil
}

func ConvertIpv4RawDataToTries(raw []model.MzrLabRpkiRtrRawData) ([]*TrieRoot, []model.MzrLabRpkiRtrRawData, error) {
	var (
		rawNormal  []model.MzrLabRpkiRtrRawData
		rawAsnZero []model.MzrLabRpkiRtrRawData
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
	ipv4RtrFulls := mzrConvertRawDataToRtrFulls(rawNormal)

	start_encode_v4 := time.Now().UnixNano()
	ipv4Tries := ConvertIpv4RtrFullsToTries(ipv4RtrFulls)
	end_encode_v4 := time.Now().UnixNano()
	encode_time_v4 := end_encode_v4 - start_encode_v4
	fmt.Printf("ipv4 encode time : %v \n", encode_time_v4)
	//fmt.Printf("ipv4 zero count : %v \n ipv4 count : %v \n", v4_zero_count, v4_count)
	f1, err := os.OpenFile("/home/cnic/rpki/newtest/compress/current/ipv4_encode_time.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f1.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f1.WriteString(fmt.Sprintf("%v \n", encode_time_v4))

	return ipv4Tries, rawAsnZero, nil
}

func ConvertIpv6RawDataToTries(raw []model.MzrLabRpkiRtrRawData) ([]*TrieRoot, []model.MzrLabRpkiRtrRawData, error) {
	var (
		rawNormal  []model.MzrLabRpkiRtrRawData
		rawAsnZero []model.MzrLabRpkiRtrRawData
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
	ipv6RtrFulls := mzrConvertRawDataToRtrFulls(rawNormal)

	start_encode_v6 := time.Now().UnixNano()
	ipv6Tries := ConvertIpv6RtrFullsToTries(ipv6RtrFulls)
	end_encode_v6 := time.Now().UnixNano()
	encode_time_v6 := end_encode_v6 - start_encode_v6
	fmt.Printf("ipv6 encode time : %v \n", encode_time_v6)
	//fmt.Printf("ipv6 zero count : %v \nipv6 count : %v \n", v6_zero_count, v6_count)
	f, err := os.OpenFile("/home/cnic/rpki/newtest/compress/current/ipv6_encode_time.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f.WriteString(fmt.Sprintf("%v \n", encode_time_v6))

	return ipv6Tries, rawAsnZero, nil
}

func ConvertTriesToPduModels(ipv4Tries, ipv6Tries []*TrieRoot) []rtrmodel.RtrPduModel {
	var rtrPduModels []rtrmodel.RtrPduModel
	rtr_count_v4 := 0
	rtr_count_v6 := 0
	if len(ipv4Tries) > 0 {

		ipv4RtrPduModels := compressConvertIpv4TriesToRtrPduModel(ipv4Tries)
		for _, m := range ipv4RtrPduModels {
			rtrPduModels = append(rtrPduModels, m)
			rtr_count_v4 += 1
		}
	}
	if len(ipv6Tries) > 0 {

		ipv6RtrPduModels := compressConvertIpv6TriesToRtrPduModel(ipv6Tries)

		for _, m := range ipv6RtrPduModels {
			rtrPduModels = append(rtrPduModels, m)
			rtr_count_v6 += 1
		}

	}
	fmt.Printf("ipv4 rtr count : %v \n", rtr_count_v4)
	fmt.Printf("ipv6 rtr count : %v \n", rtr_count_v6)
	return rtrPduModels
}

func ConvertIpv4RtrFullsToTries(ipv4RtrFulls []*RtrFull) []*TrieRoot {
	var ipv4Tries []*TrieRoot
	for _, rtrFull := range ipv4RtrFulls {
		ipv4Root := &TrieRoot{
			Asn:  uint32(rtrFull.Asn),
			Root: &TrieNode{},
		}
		for _, prefix := range rtrFull.PrefixList {
			address, _ := strconv.ParseUint(prefix.Address, 16, 32)
			InsertIpv4Prefix(ipv4Root.Root, 0, uint32(address), uint8(prefix.PrefixLength))
		}
		CalculateTrieMaxLength(ipv4Root.Root)
		ipv4Tries = append(ipv4Tries, ipv4Root)
	}
	return ipv4Tries
}

func compressConvertIpv4TriesToRtrPduModel(ipv4Tries []*TrieRoot) []*rtrmodel.RtrIpv4PrefixModel {
	var allRtrPduModel []*rtrmodel.RtrIpv4PrefixModel
	for _, trie := range ipv4Tries {
		getRtrPduModelsFromIpv4Trie(trie.Root, 0, 0, trie.Asn, &allRtrPduModel)
	}

	return allRtrPduModel
}

func getRtrPduModelsFromIpv4Trie(cur *TrieNode, depth uint8, prefix uint32, asn uint32, allRtrPduModel *[]*rtrmodel.RtrIpv4PrefixModel) {
	if cur.Value {
		wr := bytes.NewBuffer([]byte{})
		binary.Write(wr, binary.BigEndian, prefix)
		ipv4AddressSlice := wr.Bytes()
		ipv4AddressBytes := [4]byte{}
		copy(ipv4AddressBytes[0:4], ipv4AddressSlice[:])
		*allRtrPduModel = append(*allRtrPduModel, rtrmodel.NewRtrIpv4PrefixModel(1, 1, depth, cur.MaxLength, ipv4AddressBytes, asn))
	}
	if cur.Child[0] != nil {
		getRtrPduModelsFromIpv4Trie(cur.Child[0], depth+1, prefix, asn, allRtrPduModel)
	}
	if cur.Child[1] != nil {
		getRtrPduModelsFromIpv4Trie(cur.Child[1], depth+1, prefix|(1<<(31-depth)), asn, allRtrPduModel)
	}
}

func ConvertIpv6RtrFullsToTries(ipv6RtrFulls []*RtrFull) []*TrieRoot {
	var ipv6Tries []*TrieRoot
	for _, rtrFull := range ipv6RtrFulls {
		ipv6Root := &TrieRoot{
			Asn:  uint32(rtrFull.Asn),
			Root: &TrieNode{},
		}
		for _, prefix := range rtrFull.PrefixList {
			address := uint128.FromString(prefix.Address)
			InsertIpv6Prefix(ipv6Root.Root, 0, address, uint8(prefix.PrefixLength))
		}
		CalculateTrieMaxLength(ipv6Root.Root)
		ipv6Tries = append(ipv6Tries, ipv6Root)
	}
	return ipv6Tries
}

func compressConvertIpv6TriesToRtrPduModel(ipv6Tries []*TrieRoot) []*rtrmodel.RtrIpv6PrefixModel {
	var allRtrPduModel []*rtrmodel.RtrIpv6PrefixModel
	for _, trie := range ipv6Tries {
		getRtrPduModelsFromIpv6Trie(trie.Root, 0, uint128.NewUint128(0), trie.Asn, &allRtrPduModel)
	}
	return allRtrPduModel
}
func getRtrPduModelsFromIpv6Trie(cur *TrieNode, depth uint8, prefix uint128.Uint128, asn uint32, allRtrPduModel *[]*rtrmodel.RtrIpv6PrefixModel) {
	if cur.Value {
		ipv6AddressSlice := prefix.Bytes()
		ipv6AddressBytes := [16]byte{}
		copy(ipv6AddressBytes[0:16], ipv6AddressSlice[:])
		*allRtrPduModel = append(*allRtrPduModel, rtrmodel.NewRtrIpv6PrefixModel(1, 1, depth, cur.MaxLength, ipv6AddressBytes, asn))
	}
	if cur.Child[0] != nil {
		getRtrPduModelsFromIpv6Trie(cur.Child[0], depth+1, prefix, asn, allRtrPduModel)
	}
	if cur.Child[1] != nil {
		//getRtrPduModelsFromIpv6Trie(cur.Child[1], depth + 1, big.NewInt(1).Or(prefix, big.NewInt(1).Lsh(big.NewInt(1), uint(127 - depth))), asn, allRtrPduModel)
		getRtrPduModelsFromIpv6Trie(cur.Child[1], depth+1, prefix.Or(uint128.NewUint128(1).Lsh(uint(127-depth))), asn, allRtrPduModel)
	}
}

type TrieNode struct {
	Value     bool
	Child     [2]*TrieNode
	MaxLength uint8
}

type TrieRoot struct {
	Asn  uint32
	Root *TrieNode
}

func InsertIpv4Prefix(cur *TrieNode, depth uint8, address uint32, prefixLength uint8) {
	if depth == prefixLength {
		cur.Value = true
		cur.MaxLength = depth
	} else {
		i := (address >> (31 - depth)) & 1
		if cur.Child[i] == nil {
			cur.Child[i] = &TrieNode{}
		}
		InsertIpv4Prefix(cur.Child[i], depth+1, address, prefixLength)
	}
}

func InsertIpv6Prefix(cur *TrieNode, depth uint8, address uint128.Uint128, prefixLength uint8) {
	if depth == prefixLength {
		cur.Value = true
		cur.MaxLength = depth
	} else {
		i := uint128.NewUint128(1).And(address.Rsh(uint(127 - depth))).Uint64()
		//i := big.NewInt(1).And(big.NewInt(1), big.NewInt(1).Rsh(address, uint(127-depth))).Uint64()
		if cur.Child[i] == nil {
			cur.Child[i] = &TrieNode{}
		}
		InsertIpv6Prefix(cur.Child[i], depth+1, address, prefixLength)
	}
}

func CalculateTrieMaxLength(cur *TrieNode) {
	if cur.Child[0] != nil {
		CalculateTrieMaxLength(cur.Child[0])
	}
	if cur.Child[1] != nil {
		CalculateTrieMaxLength(cur.Child[1])
	}
	if cur.Value == true {
		if cur.Child[0] != nil && cur.Child[0].Value == true && cur.Child[1] != nil && cur.Child[1].Value == true {
			if cur.Child[0].MaxLength <= cur.Child[1].MaxLength {
				cur.MaxLength = cur.Child[0].MaxLength
			} else {
				cur.MaxLength = cur.Child[1].MaxLength
			}
			if cur.MaxLength == cur.Child[0].MaxLength {
				cur.Child[0].Value = false
				//cur.Child[0] = nil
			}
			if cur.MaxLength == cur.Child[1].MaxLength {
				cur.Child[1].Value = false
				//cur.Child[1] = nil
			}
		}
	}
}
