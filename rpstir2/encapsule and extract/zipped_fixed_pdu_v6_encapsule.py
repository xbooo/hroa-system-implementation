inputFile = open("m_pdu_v6_fur_0712.txt", "r")
outputFile = open("output_" + inputFile.name, "w")
print("Input File:", inputFile.name)
print("Output File:", outputFile.name)


keyMask = [0] * 129

for i in range(0, 129):
    keyMask[i] = (1 << 128) - (1 << (128 - i // 5 * 5))
#    print(i, i // 5 * 5, "{:b}".format(keyMask[i]))


totalPrefixCount = 0
encapsuledPduCount = 0
asnCount = 0

while True:
    line = inputFile.readline()
    if not line:
        break

    lineSplit = line.split()
    if lineSplit[0][0] == '{':
        continue
    asn = int(lineSplit[0])
    asnCount += 1
    prefixCount = int(lineSplit[1])
    prefixDict = {}
    flag = 1
    totalPrefixCount += prefixCount
    
    for i in range(0, prefixCount):
        prefixAddr = int(lineSplit[2 + 2 * i], 16)
        prefixLength = int(lineSplit[3 + 2 * i])
        prefixKey = prefixAddr & keyMask[prefixLength]
        subtreePath = ((prefixAddr ^ prefixKey)>> (128 - prefixLength))
        subtreeMask = 1 << ((1 << (prefixLength % 5)) + subtreePath)
        dictKey = (1 << prefixLength // 5 * 5) + (prefixKey >> (128 - (prefixLength // 5 * 5)))
        if dictKey in prefixDict:
            prefixDict[dictKey] |= subtreeMask
        else:
            prefixDict[dictKey] = subtreeMask
#        print("{:0>128b}\n{:b}".format(prefixAddr, keyMask[prefixLength]))

    encapsuledPduCount += len(prefixDict)
    outputFile.write("{} {}".format(len(prefixDict) * 20 + 12, asn))
    for key, value in prefixDict.items():
        outputFile.write(" {:0>32x} {:0>8x}".format(key, value | flag))
#        print("{:0>32b}".format(key))
    outputFile.write("\n")

#        print("{:0>128b}".format(key))

inputFile.close()
outputFile.close()
print("Total Prefix Count:", totalPrefixCount)
print("Encapsuled PDU Count:", encapsuledPduCount)
print("Total Prefix PDU Size:", totalPrefixCount * 32)
print("Encapsuled PDU Size:", encapsuledPduCount * 32)
print("Zipped PDU Count:", asnCount)
print("Zipped PDU Size:", encapsuledPduCount * 20 + asnCount * 12)
