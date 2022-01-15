def prefixMask(length):
    return (1 << 128) - (1 << (128 - length))

def countNum(pList, minLen, maxLen):
    qList = []
    for p in pList:
        if p[1] >= minLen and p[1] <= maxLen:
            qList.append(p[0] >> (128 - minLen))
    return len(set(qList))

def calcuSize(x, y):
    if(y - x - 2 >= 0):
        bitmapSize = 1 << (y - x - 2)
    else:
        bitmapSize = 1
    return 28 + bitmapSize

inputFile = open("m_pdu_v6_cur_0712.txt", "r")
outputFile = open("output_" + inputFile.name, "w")

print("Input File:", inputFile.name)
print("Output File:", outputFile.name)

totalPrefixCount = 0
totalFlexPduSize = 0
totalFlexPduCount = 0
totalEncapsuledPduCount = 0
while True:
    line = inputFile.readline()
    if not line:
        break

    lineSplit = line.split()
    if lineSplit[0][0] == '{':
        continue
    asn = int(lineSplit[0])
    prefixCount = int(lineSplit[1])
    prefixList = []
    prefixDict = {}
    flag = 1
    
    for i in range(0, prefixCount):
        prefixAddr = int(lineSplit[2 + 2 * i], 16)
        prefixLength = int(lineSplit[3 + 2 * i])
        prefixList.append([prefixAddr, prefixLength])

    qList = []
    cNum = []
    for i in range(0, 129):
        qList.append([])
    for i in range(0, prefixCount):
        qList[prefixList[i][1]].append(prefixList[i][0])
    for i in range(0, 129):
        rList = []
        cList = []
        for j in range(0, 129):
            if j < i:
                cList.append(0)
                continue
            for pfx in qList[j]:
                rList.append(pfx >> (128 - i))
            cList.append(len(set(rList)))
        cNum.append(cList)
            
    dp = [0] * 129
    opt = [0] * 129
    hang = [0] * 129
    nxt = [0] * 129
    for i in range(0, 129):
        dp[i] = calcuSize(0, i) * cNum[0][i]
        opt[i] = -1
        for j in range(0, i if i < 128 else i - 1):
            tmp = dp[j] + calcuSize(j + 1, i) * cNum[j + 1][i]
            if(tmp < dp[i]):
                dp[i] = tmp
                opt[i] = j

    flexPduCount = 0
    cur = 128
    while(cur >= 0):
        flexPduCount += cNum[opt[cur] + 1][cur]
        for i in range(cur, opt[cur], -1):
            hang[i] = opt[cur] + 1
        tmp = cur - opt[cur]
        nxt[opt[cur] + 1] = cur
        cur = opt[cur]

    totalFlexPduCount += flexPduCount
    totalPrefixCount += prefixCount
    totalFlexPduSize += dp[128]

    for i in range(0, prefixCount):
        prefixAddr = prefixList[i][0]
        prefixLength = prefixList[i][1]
        hangLen = hang[prefixLength]
        prefixKey = prefixAddr & prefixMask(hangLen)
        subtreePath = ((prefixAddr ^ prefixKey)>> (128 - prefixLength))
        subtreeMask = (1 << ((1 << (prefixLength - hangLen)) + subtreePath))
        dictKey = (1 << hangLen) + (prefixKey >> (128 - hangLen))
        if dictKey in prefixDict:
            prefixDict[dictKey] |= subtreeMask
        else:
            prefixDict[dictKey] = subtreeMask
 
    totalEncapsuledPduCount += len(prefixDict)
    
    for key, value in prefixDict.items():
        keyLength = len(f'{key:b}') - 1
        outputStr = "{} {} {:0>32x} {:0>" + str(2 * (calcuSize(keyLength, nxt[keyLength]) - 28)) +"x}\n"
        outputFile.write(outputStr.format(calcuSize(keyLength, nxt[keyLength]), asn, key, value | flag))


inputFile.close()
outputFile.close()
print("Flex PDU Count:", totalFlexPduCount)
print("Flex PDU Size:", totalFlexPduSize)
print("Total Prefix Count:", totalPrefixCount)
print("Total Encapsuled PDU Count:", totalEncapsuledPduCount)

