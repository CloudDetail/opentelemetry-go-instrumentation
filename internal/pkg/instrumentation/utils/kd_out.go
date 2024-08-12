package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
)

const TxLogEnvVar = "KD_OUTPUT_LOG"

var (
	lookupEnv     = os.LookupEnv
	pidServiceMap sync.Map
	bootTime      int64  = 0
	devNull       Writer = newWriter()
)

func CachePidServiceName(pid uint32, svcName string) {
	pidServiceMap.Store(pid, svcName)
}

func SetBootTime(bt int64) {
	bootTime = bt
}

func WriteKdClientEvent(traceId string, startTime int64, endTime int64, pid uint32) {
	result := fmt.Sprintf("kd-txnet@1!%s!%d!%d!%d!",
		traceId, startTime+bootTime, endTime+bootTime, pid)
	devNull.Write(result)
}

func WriteKdServerEnter(traceId string, startTime int64, topSpan bool, url string, goId uint64, pid uint32) {
	topFlag := 0
	if topSpan {
		topFlag = 1
	}
	result := fmt.Sprintf("kd-txin@9!%s!%d!%d!%d!%s!%d!%d!",
		traceId, startTime+bootTime, 1, topFlag, url, pid, goId)
	devNull.Write(result)
}

func WriteKdServerExit(traceId string, endTime int64, hasError bool, spanId string, isSampled bool, goId uint64, pid uint32) {
	errorFlag := 0
	if hasError {
		errorFlag = 1
	}
	sampleFlag := 0
	if isSampled {
		sampleFlag = 1
	}
	serviceName := "???"
	if svcNameInterface, exist := pidServiceMap.Load(pid); exist {
		serviceName = svcNameInterface.(string)
	}
	result := fmt.Sprintf("kd-txout@9!%s!%d!%s!%d!%d!%s!%s!%d!!0!%d!%d!",
		traceId, endTime+bootTime, "otel", 0, errorFlag, spanId, serviceName, sampleFlag, pid, goId)
	devNull.Write(result)
}

type Writer interface {
	Write(data string)
}

type LogWriter struct {
}

func (writer *LogWriter) Write(data string) {
	log.Println(data)
}

type DevWriter struct {
	file *os.File
}

func NewDevWriter() *DevWriter {
	devNull, err := os.OpenFile("/dev/null", os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("Error to write DevNull, %s\n", err.Error())
		return nil
	}

	return &DevWriter{
		file: devNull,
	}
}

func (devNull *DevWriter) Write(data string) {
	buffer := bufio.NewWriter(devNull.file)
	if _, err := buffer.WriteString(data); err == nil {
		buffer.Flush()
	}
}

func newWriter() Writer {
	if log, ok := lookupEnv(TxLogEnvVar); ok && log == "true" {
		return &LogWriter{}
	}

	if devWriter := NewDevWriter(); devWriter != nil {
		return devWriter
	}
	return &LogWriter{}
}
