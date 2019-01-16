package engine

import (
	"fmt"

	"github.com/bulklog/bulklog/output"
	"github.com/gomodule/redigo/redis"
)

func getRedisPipeoutputs(red *redis.Pool, pipeKey string, outputs map[string]output.Interface) (remainingoutputs map[string]output.Interface, err error) {
	conn := red.Get()
	defer conn.Close()
	key := fmt.Sprintf("%s.outputs", pipeKey)
	remainingoutputsLen, err := conn.Do("LLen", key)
	if err != nil {
		return nil, fmt.Errorf("(LLEN pipeKey.outputs).%s", err)
	}
	if remainingoutputsLen == 0 {
		return map[string]output.Interface{}, nil
	}
	remainingoutputNamesI, err := conn.Do("LRANGE", key, 0, remainingoutputsLen)
	if err != nil {
		return nil, fmt.Errorf("(LRANGE pipeKey.outputs).%s", err)
	}
	remainingoutputNames := remainingoutputNamesI.([]interface{})
	remainingoutputs = make(map[string]output.Interface)
	var (
		outputNameI interface{}
		outputName  string
	)
	for _, outputNameI = range remainingoutputNames {
		outputName = string(outputNameI.([]byte))
		if cons, ok := outputs[outputName]; ok {
			remainingoutputs[outputName] = cons
		}
	}
	return remainingoutputs, nil
}

func addRedisPipeoutputs(conn redis.Conn, pipeKey string, outputs map[string]output.Interface) (err error) {
	key := fmt.Sprintf("%s.outputs", pipeKey)
	args := make([]interface{}, 0, len(outputs)+1)
	args = append(args, key)
	var outputName string
	for outputName = range outputs {
		args = append(args, outputName)
	}
	err = conn.Send("RPUSH", args...)
	if err != nil {
		return fmt.Errorf("(RPUSH pipeKey.outputs outputNames...).%s", err)
	}
	return
}

func deleteRedisPipeoutput(red *redis.Pool, pipeKey, outputName string) (err error) {
	conn := red.Get()
	defer conn.Close()
	_, err = conn.Do("LREM", fmt.Sprintf("%s.outputs", pipeKey), 0, outputName)
	if err != nil {
		return fmt.Errorf("(LREM pipeKey.outputs outputName).%s", err)
	}
	return nil
}

func deleteRedisPipeoutputs(conn redis.Conn, pipeKey string) (err error) {
	err = conn.Send("DEL", fmt.Sprintf("%s.outputs", pipeKey))
	if err != nil {
		return fmt.Errorf("(DEL pipeKey.outputs).%s", err)
	}
	return nil
}
