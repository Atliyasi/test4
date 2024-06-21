package main

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/baidubce/bce-sdk-go/services/bos"
)

const (
	MAX_BYTE_IN_MEM = 256 * 1024
	MAX_FAIL_TIMES  = 3
)

const (
	AK           = "7c23300872e24f1c9da80c6ac81c4f7b"
	ENCRYPT_SK   = "yW7BmI1c2DuWhWbguoV2Usg4ZEhB+femfn50Ja/LY9wtIigf5Z4TQNdYK8QP5nIF"
	ENCRYPT_SK_2 = "yW7BmI1c2DuWhWbguoV2Usg4ZEhBBBBB+femfn50Ja/LY9wtIigf5Z4TQNdYK8QP5nIF"
)

// 解密后的SK
var SK string

type Counter struct {
	mutex   sync.Mutex
	counter int64
}

type BosHandler struct {
	bosClient  *bos.Client
	filename   string
	storePath  string
	commandBuf []byte
	// 数据上报BOS改为追加上传的方式,避免占用过多的用户内存和上传数据时的网络带宽
	// 每当内存中存储的命令超过256KB之后，将数据上传至BOS
	// offset存储下一次追加数据的偏移量，初始为0
	offset int64
	// 记录上报数据时，连续失败的次数
	failTimes int
}

var globalCount = Counter{counter: 0}

func GenerateAuditFilename() string {
	// generate audit filename
	current := time.Now()

	globalCount.mutex.Lock()
	defer globalCount.mutex.Unlock()
	globalCount.counter++

	return current.Format("20060102150405") + "_" + strconv.FormatInt(globalCount.counter, 10)
}

func GenerateStorePath() string {
	// generate store path
	// bucket_name/ipaddr
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	readAddr := ""
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				readAddr = ipnet.IP.String()
				break
			}
		}
	}
	return readAddr
}

func NEW() (*BosHandler, error) {
	clientConfig := bos.BosClientConfiguration{
		Ak:               AK,
		Sk:               SK,
		Endpoint:         EndPoint,
		RedirectDisabled: false,
	}

	// create bosclient
	bosHandler := &BosHandler{bosClient: nil}
	bosClient, err := bos.NewClientWithConfig(&clientConfig)
	if err != nil {
		return bosHandler, err
	}

	bosHandler.bosClient = bosClient
	bosHandler.filename = GenerateAuditFilename()
	bosHandler.storePath = GenerateStorePath()
	return bosHandler, nil
}

func (bosHandler *BosHandler) AppendStringToBuffer(p []byte, n int) error {
	// put command to buffer
	bosHandler.commandBuf = append(bosHandler.commandBuf, p[0:n]...)
	if len(bosHandler.commandBuf) > MAX_BYTE_IN_MEM {
		// 内存中的数据量超过256KB，将数据追加到BOS中
		newOffset, err := bosHandler.FlushStringToBOS()
		if err != nil {
			bosHandler.failTimes++
			if bosHandler.failTimes >= MAX_FAIL_TIMES {
				// 连续失败的次数超过3次之后
				// 将缓存中的数据清空，放弃当前审计数据的上报
				bosHandler.failTimes = 0
				bosHandler.commandBuf = bosHandler.commandBuf[0:0]
			}
			return err
		}

		bosHandler.failTimes = 0
		bosHandler.offset = newOffset
		bosHandler.commandBuf = bosHandler.commandBuf[0:0]
		return nil
	}
	return nil
}

func (bosHandler *BosHandler) FlushStringToBOS() (int64, error) {
	// put string to BOS
	// path: bucketname/path/filename
	// eg: ensec-gotty-log/192.168.1.10/filename
	var storePath string
	if len(bosHandler.storePath) == 0 {
		storePath = bosHandler.filename
	} else {
		storePath = bosHandler.storePath + "/" + bosHandler.filename
	}

	// 以追加的方式上报数据
	res, err := bosHandler.bosClient.SimpleAppendObjectFromString(BucketName, storePath,
		string(bosHandler.commandBuf), bosHandler.offset)
	return res.NextAppendOffset, err
}

func (bosHandler *BosHandler) DataAvailable() bool {
	return len(bosHandler.commandBuf) > 0
}

func (bosHandler *BosHandler) DestoryBosHandler() {
	bosHandler.commandBuf = nil
	bosHandler.bosClient = nil
}
