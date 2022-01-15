inputFile = open("output_m_pdu_v4_cur_0712.txt", "r")
outputFile = open("extract_" + inputFile.name, "w")
print("Input File:", inputFile.name)
print("Output File:", outputFile.name)

encapsuledPduCount = 0
totalPrefixCount = 0

while True:
    line = inputFile.readline()
    if not line:
        break
    length, asn, xprefix, xbitmap = line.split()
    encapsuledPduCount += 1
    length = int(length)
    keyPrefix = int(xprefix, 16)
    bitmap = int(xbitmap, 16)
    keyPrefixLength = len(f'{keyPrefix:b}') - 1
    keyPrefix ^= 1 << keyPrefixLength
    keyPrefix <<= 32 - keyPrefixLength
    for i in range(1, (length - 16) * 8):
        if((bitmap >> i) & 1):
            subPrefixLength = len(f'{i:b}') - 1
            path = i ^ (1 << subPrefixLength)
            prefixLength = keyPrefixLength + subPrefixLength
            outputFile.write("{} {:0>8x} {}\n".format(asn, keyPrefix + (path << (32 - prefixLength)), prefixLength))
            totalPrefixCount += 1

inputFile.close()
outputFile.close()

print("Encapsuled PDU Count:", encapsuledPduCount)
print("Extrated Prefix Count:", totalPrefixCount)
