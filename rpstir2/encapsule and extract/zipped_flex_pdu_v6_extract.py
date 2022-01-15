inputFile = open("output_m_pdu_v6_cur_0712.txt", "r")
outputFile = open("extract_" + inputFile.name, "w")
print("Input File:", inputFile.name)
print("Output File:", outputFile.name)

encapsuledPduCount = 0
totalPrefixCount = 0

while True:
    line = inputFile.readline()
    if not line:
        break
    lineSplit = line.split()
    PduLength = int(lineSplit[0])
    asn = int(lineSplit[1])
    bcount = int(lineSplit[2])
    encapsuledPduCount += 1

    for k in range(0, bcount):
        xprefix = lineSplit[3 + k * 3]
        bitmapSize = 1 << int(lineSplit[4 + k * 3])
        xbitmap = lineSplit[5 + k * 3]
        keyPrefix = int(xprefix, 16)
        bitmap = int(xbitmap, 16)
        keyPrefixLength = len(f'{keyPrefix:b}') - 1
        keyPrefix ^= 1 << keyPrefixLength
        keyPrefix <<= 128 - keyPrefixLength
#        print(bitmapSize)
        for i in range(1, bitmapSize * 2):
            if((bitmap >> i) & 1):
                subPrefixLength = len(f'{i:b}') - 1
                path = i ^ (1 << subPrefixLength)
                prefixLength = keyPrefixLength + subPrefixLength
                outputFile.write("{} {:0>32x} {}\n".format(asn, keyPrefix + (path << (128 - prefixLength)), prefixLength))
                totalPrefixCount += 1

inputFile.close()
outputFile.close()

print("Encapsuled PDU Count:", encapsuledPduCount)
print("Extrated Prefix Count:", totalPrefixCount)
