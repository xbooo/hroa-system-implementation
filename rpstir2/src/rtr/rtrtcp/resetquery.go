package rtrtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"model"
	"os"
	db "rtr/db"
	rtrmodel "rtr/model"
	"time"
	"uint128"

	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
)

func ParseToResetQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
	var zero16 uint16
	var length uint32

	// get zero16
	err = binary.Read(buf, binary.BigEndian, &zero16)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get zero fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get length fail: ", buf, err)
		rtrError := rtrmodel.NewRtrError(
			err,
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToResetQuery():PDU_TYPE_RESET_QUERY,  length must be 8 ", buf, length)
		rtrError := rtrmodel.NewRtrError(
			errors.New("pduType is SERIAL QUERY, length must be 8"),
			true, protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	rq := rtrmodel.NewRtrResetQueryModel(protocolVersion)
	belogs.Debug("ParseToResetQuery():get PDU_TYPE_RESET_QUERY ", buf, jsonutil.MarshalJson(rq))
	return rq, nil
}

func ProcessResetQuery(rtrPduModel rtrmodel.RtrPduModel) (resetResponses []rtrmodel.RtrPduModel, err error) {
	start_database_v4 := time.Now().UnixNano()
	ipv4Raw, sessionId, serialNumber, err := db.MzrGetIpv4RtrFullRawAndSessionIdAndSerialNumber()
	end_database_v4 := time.Now().UnixNano()
	database_v4 := end_database_v4 - start_database_v4

	if err != nil {
		belogs.Error("ProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}

	start_database_v6 := time.Now().UnixNano()
	ipv6Raw, sessionId, serialNumber, err := db.MzrGetIpv6RtrFullRawAndSessionIdAndSerialNumber()
	end_database_v6 := time.Now().UnixNano()
	database_v6 := end_database_v6 - start_database_v6
	fmt.Printf("ipv4 database time : %v\nipv6 database time : %v\n", database_v4, database_v6)

	f1, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_troa_current_ipv4.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f1.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f1.WriteString(fmt.Sprintf("%v \n", database_v4))

	f, err := os.OpenFile("/home/cnic/rpki/newtest/database/cst_troa_current_ipv6.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	defer f.Close()
	if err != nil {
		fmt.Printf("openfile wrong !!!!\n")
	}
	f.WriteString(fmt.Sprintf("%v \n", database_v6))

	if err != nil {
		belogs.Error("ProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}

	pduModels := ConvertRawDataToPduModels(ipv4Raw, rtrPduModel.GetProtocolVersion())
	pduModels = append(pduModels, ConvertRawDataToPduModels(ipv6Raw, rtrPduModel.GetProtocolVersion())...)

	belogs.Debug("ProcessResetQuery(): rtrFulls, sessionId, serialNumber: ", len(pduModels), sessionId, serialNumber)
	if err != nil {
		belogs.Error("ProcessResetQuery(): ConvertRtrFullsToRtrPduModels fail: ", err)
		return resetResponses, err
	}

	rtrPduModels, err := assembleResetResponses(pduModels, rtrPduModel.GetProtocolVersion(), sessionId, serialNumber)
	if err != nil {
		belogs.Error("ProcessResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}
	return rtrPduModels, nil
}

func ConvertRawDataToPduModels(raw []model.MzrLabRpkiRtrRawData, protocolVersion uint8) []rtrmodel.RtrPduModel {
	var res []rtrmodel.RtrPduModel

	rtr_count_v4 := 0
	rtr_count_v6 := 0
	for _, d := range raw {
		rawAddress := uint128.FromString(d.Address)
		addressBytes := rawAddress.Bytes()
		if rawAddress.High == 0 && rawAddress.Low < (1<<32) {
			ipv4 := [4]byte{}
			copy(ipv4[:], addressBytes[12:16])
			res = append(res, rtrmodel.NewRtrIpv4PrefixModel(protocolVersion, 1, uint8(d.PrefixLength), uint8(d.MaxLength), ipv4, uint32(d.Asn)))
			rtr_count_v4 += 1
		} else {
			ipv6 := [16]byte{}
			copy(ipv6[:], addressBytes[:])
			res = append(res, rtrmodel.NewRtrIpv6PrefixModel(protocolVersion, 1, uint8(d.PrefixLength), uint8(d.MaxLength), ipv6, uint32(d.Asn)))
			rtr_count_v6 += 1
		}
	}
	fmt.Printf("ipv4 rtr count : %v \n", rtr_count_v4)
	fmt.Printf("ipv6 rtr count : %v \n", rtr_count_v6)
	return res
}

//func ConvertRtrFullsToRtrPduModels(rtrFulls []model.LabRpkiRtrFull, protocolVersion uint8) ([]rtrmodel.RtrPduModel, error) {
//	var rtrPduModels []rtrmodel.RtrPduModel
//	for i, _ := range rtrFulls {
//		rtrPduModel, err := convertRtrFullToRtrPduModel(&rtrFulls[i], protocolVersion)
//		if err != nil {
//			belogs.Error("assembleResetResponses(): convertRtrFullToRtrPduModel fail: ", err)
//			return rtrPduModels, err
//		}
//		belogs.Debug("assembleResetResponses(): rtrPduModel : ", jsonutil.MarshalJson(rtrPduModel))
//		rtrPduModels = append(rtrPduModels, rtrPduModel)
//	}
//	return rtrPduModels, nil
//}
// when len(rtrFull)==0, it is an error with no_data_available
func assembleResetResponses(middleRtrPduModels []rtrmodel.RtrPduModel, protocolVersion uint8, sessionId uint16,
	serialNumber uint32) (rtrPduModels []rtrmodel.RtrPduModel, err error) {
	rtrPduModels = make([]rtrmodel.RtrPduModel, 0)

	if len(middleRtrPduModels) > 0 {
		belogs.Debug("assembleResetResponses(): will send will send Cache Response of all rtr,",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ", len(rtr): ", len(middleRtrPduModels))

		cacheResponseModel := rtrmodel.NewRtrCacheResponseModel(protocolVersion, sessionId)
		belogs.Debug("assembleResetResponses(): cacheResponseModel : ", jsonutil.MarshalJson(cacheResponseModel))

		rtrPduModels = append(rtrPduModels, cacheResponseModel)

		rtrPduModels = append(rtrPduModels, middleRtrPduModels...)

		endOfDataModel := assembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
		belogs.Debug("assembleResetResponses(): endOfDataModel : ", jsonutil.MarshalJson(endOfDataModel))

		rtrPduModels = append(rtrPduModels, endOfDataModel)
		belogs.Info("assembleResetResponses(): will send will send Cache Response of all rtr,",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber,
			", len(rtr): ", len(middleRtrPduModels), ",  len(rtrPduModels):", len(rtrPduModels))

	} else {
		belogs.Debug("assembleResetResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber)
		errorReportModel := rtrmodel.NewRtrErrorReportModel(protocolVersion, rtrmodel.PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)

		rtrPduModels = append(rtrPduModels, errorReportModel)
		belogs.Info("assembleResetResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			",  protocolVersion:", protocolVersion, ",   sessionId:", sessionId, ",  serialNumber:", serialNumber, ",  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels))

	}
	return rtrPduModels, nil

}

//func convertRtrFullToRtrPduModel(rtrFull *model.LabRpkiRtrFull, protocolVersion uint8) (rtrPduModel rtrmodel.RtrPduModel, err error) {
//	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrFull.Address)
//	if ipType == iputil.Ipv4Type {
//		ipv4 := [4]byte{0x00}
//		copy(ipv4[:], ipHex[:])
//		rtrIpv4PrefixModel := rtrmodel.NewRtrIpv4PrefixModel(protocolVersion, 1, uint8(rtrFull.PrefixLength),
//			uint8(rtrFull.MaxLength), ipv4, uint32(rtrFull.Asn))
//		return rtrIpv4PrefixModel, nil
//	} else if ipType == iputil.Ipv6Type {
//		ipv6 := [16]byte{0x00}
//		copy(ipv6[:], ipHex[:])
//		rtrIpv6PrefixModel := rtrmodel.NewRtrIpv6PrefixModel(protocolVersion, 1, uint8(rtrFull.PrefixLength),
//			uint8(rtrFull.MaxLength), ipv6, uint32(rtrFull.Asn))
//		return rtrIpv6PrefixModel, nil
//	}
//	return rtrPduModel, errors.New("convert to rtr format, error ipType")
//}

type Prefix struct {
	Address      string
	PrefixLength uint64
}

type RtrFull struct {
	Asn        uint64
	Length     uint
	PrefixList []Prefix
}
