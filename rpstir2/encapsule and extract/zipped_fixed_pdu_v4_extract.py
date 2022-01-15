inputFile = open("output_m_pdu_v4_fur_0712.txt", "r")
outputFile = open("extract_" + inputFile.name, "w")
print("Input File:", inputFile.name)
print("Output File:", outputFile.name)

encapsuledPduCount = 0
totalPrefixCount = 0
bLen = [0] * 1 + [1] * 1 + [2] * 2 + [3] * 4 + [4] * 8 + [5] * 16

while True:
    line = inputFile.readline()
    if not line:
        break
    lineSplit = line.split()
    length = int(lineSplit[0])
    asn = int(lineSplit[1])
    encapsuledPduCount += 1

    for k in range(0, (length - 12) // 8):
        xprefix = lineSplit[2 + k * 2]
        xbitmap = lineSplit[3 + k * 2]
        keyPrefix = int(xprefix, 16)
        bitmap = int(xbitmap, 16)
        keyPrefixLength = len(f'{keyPrefix:b}') - 1
        keyPrefix ^= 1 << keyPrefixLength
        keyPrefix <<= 32 - keyPrefixLength
        for i in range(1, 32):
            if((bitmap >> i) & 1):
                subPrefixLength = bLen[i] - 1
                path = i ^ (1 << subPrefixLength)
                prefixLength = keyPrefixLength + subPrefixLength
                outputFile.write("{} {:0>8x} {}\n".format(asn, keyPrefix + (path << (32 - prefixLength)), prefixLength))
                totalPrefixCount += 1

inputFile.close()
outputFile.close()

print("Encapsuled PDU Count:", encapsuledPduCount)
print("Extrated Prefix Count:", totalPrefixCount)
