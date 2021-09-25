package tests

import (
	"encoding/json"
	"fmt"
	"path"
	"testing"
)

func TestCSVSimple(t *testing.T) {
	fpath := path.Join(getCurrentDir(t), "topics.txt")

	cmd := fmt.Sprintf("cat %v | csv row[17:18]", fpath)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines), 3)
	assertStringEquals(lines[0], "TOPIC,PARTITION,CURRENT-OFFSET,LOG-END-OFFSET,LAG,CONSUMER-ID,HOST,CLIENT-ID")
	assertStringEquals(lines[1], "topic3,0,26984839,26984839,0,consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1/10.9.27.3,consumer-21")
	assertStringEquals(lines[2], "")

	cmd = fmt.Sprintf("cat %v | csv row[17:18] col[0,2]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 3)
	assertStringEquals(lines[0], "TOPIC,CURRENT-OFFSET")
	assertStringEquals(lines[1], "topic3,26984839")
	assertStringEquals(lines[2], "")

	cmd = fmt.Sprintf("cat %v | csv row[17:18] col[0,2] merge:::", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 3)
	assertStringEquals(lines[0], "TOPIC::CURRENT-OFFSET")
	assertStringEquals(lines[1], "topic3::26984839")
	assertStringEquals(lines[2], "")

	cmd = fmt.Sprintf("cat %v | csv row[17:18] col[0,2] out..csv", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 3)
	assertStringEquals(lines[0], "TOPIC,CURRENT-OFFSET")
	assertStringEquals(lines[1], "topic3,26984839")
	assertStringEquals(lines[2], "")

	cmd = fmt.Sprintf("cat %v | csv row[16:18] col[0,2] lmerge:--", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 2)
	assertStringEquals(lines[0], "topic2,30--topic3,26984839")
	assertStringEquals(lines[1], "")

	cmd = fmt.Sprintf("cat %v | csv row[16:18] col[0,2] lmerge:' :: ' merge:' : '", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 2)
	assertStringEquals(lines[0], "topic2 : 30 :: topic3 : 26984839")
	assertStringEquals(lines[1], "")

	cmd = fmt.Sprintf("cat %v | csv out..csv sort[2] head[topic,partition,in,out,lag,consumer,client]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "topic,partition,in,out,lag,consumer,client")
	assertStringEquals(lines[1], "topic2,2,21,21,0,consumer-2-ebe5eadf-8712-4b84-8951-afc00117e325/10.9.27.3,consumer-2")
	assertStringEquals(lines[17], "topic3,0,26984839,26984839,0,consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1/10.9.27.3,consumer-21")

	cmd = fmt.Sprintf("cat %v | csv out..csv sort[2] head[topic,partition,in,out,lag,consumer,client,calc] 'calc([2]+[3])'", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "topic,partition,in,out,lag,consumer,client,calc")
	assertStringEquals(lines[1], "topic2,2,21,21,0,consumer-2-ebe5eadf-8712-4b84-8951-afc00117e325/10.9.27.3,consumer-2,42")
	assertStringEquals(lines[17], "topic3,0,26984839,26984839,0,consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1/10.9.27.3,consumer-21,53969678")

	cmd = fmt.Sprintf("cat %v | csv out..table sort[2] head[topic,partition,in,out,lag,consumer,client,calc] 'calc([2]+[3])'", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "topic     partition    in          out         lag    consumer                                                      client         calc        ")
	assertStringEquals(lines[1], "topic2    2            21          21          0      consumer-2-ebe5eadf-8712-4b84-8951-afc00117e325/10.9.27.3     consumer-2     42          ")
	assertStringEquals(lines[17], "topic3    0            26984839    26984839    0      consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1/10.9.27.3    consumer-21    53969678    ")

	//todo sort

}

func TestCSVColTransform(t *testing.T) {
	fpath := path.Join(getCurrentDir(t), "topics.txt")

	cmd := fmt.Sprintf("cat %v | csv tr..c5..split:/..col[-1] out..csv", fpath)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "TOPIC,PARTITION,CURRENT-OFFSET,LOG-END-OFFSET,LAG,CONSUMER-ID,HOST,CLIENT-ID")
	assertStringEquals(lines[1], "topic1,44,808699,808699,0,10.9.27.3,consumer-5")
	assertStringEquals(lines[17], "topic3,0,26984839,26984839,0,10.9.27.3,consumer-21")

	cmd = fmt.Sprintf("cat %v | csv  col[5,1] tr..c5..split:/..col[-1] out..csv", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "CONSUMER-ID,PARTITION")
	assertStringEquals(lines[1], "10.9.27.3,44")
	assertStringEquals(lines[17], "10.9.27.3,0")

	cmd = fmt.Sprintf("cat %v | csv col[5,0] tr..c5..split:/..col[0,1]..merge::: out..csv", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "CONSUMER-ID,TOPIC")
	assertStringEquals(lines[1], "consumer-5-c6ac0ffe-b453-41ad-a3a4-2b1265735ed3::10.9.27.3,topic1")
	assertStringEquals(lines[17], "consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1::10.9.27.3,topic3")

	cmd = fmt.Sprintf("cat %v | csv col[5,0] tr..c5..split:/..col[0,1]..merge::: out..table", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "CONSUMER-ID                                                    TOPIC     ")
	assertStringEquals(lines[1], "consumer-5-c6ac0ffe-b453-41ad-a3a4-2b1265735ed3::10.9.27.3     topic1    ")
	assertStringEquals(lines[17], "consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1::10.9.27.3    topic3    ")

	//chained col transform
	cmd = fmt.Sprintf("cat %v | csv  col[5,1] tr..c5..split:/..col[-1] tr#c5#split:.#merge:: out..csv", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "CONSUMER-ID,PARTITION")
	assertStringEquals(lines[1], "10:9:27:3,44")
	assertStringEquals(lines[17], "10:9:27:3,0")

	//tr prefix and suffix
	cmd = fmt.Sprintf("cat %v | csv out..csv col[5,1] tr..c5..split:/..col[-1]..pfx:'<<<'..sfx:'>>>' out..csv", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "CONSUMER-ID,PARTITION")
	assertStringEquals(lines[1], "<<<10.9.27.3>>>,44")
	assertStringEquals(lines[17], "<<<10.9.27.3>>>,0")

	//ncol
	cmd = fmt.Sprintf("cat %v | csv  col[5,1] tr..c5..split:/..ncol[0] out..csv", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "CONSUMER-ID,PARTITION")
	assertStringEquals(lines[1], "10.9.27.3,44")
	assertStringEquals(lines[17], "10.9.27.3,0")
}

func TestCSVColGroupAndSort(t *testing.T) {
	fpath := path.Join(getCurrentDir(t), "topics.txt")

	// group by  and sum values
	cmd := fmt.Sprintf("cat %v | csv col[2,3,4,6] group[3] sort[1,0]", fpath)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "HOST,CURRENT-OFFSET,LOG-END-OFFSET,LAG")
	assertStringEquals(lines[1], "consumer-3,50,50,0")
	assertStringEquals(lines[7], "consumer-21,26984839,26984839,0")

	// string merge cols
	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0] sort[0,2] out..table", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "TOPIC     HOST           CURRENT-OFFSET    LOG-END-OFFSET    LAG    ")
	assertStringEquals(lines[1], "topic1    consumer-5     6468718           6468718           0      ")
	assertStringEquals(lines[2], "          consumer-6                                                ")
	assertStringEquals(lines[7], "topic3    consumer-21    26984839          26984839          0      ")

	// string merge cols
	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0] sort[0,2] out..csv", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 5)
	assertStringEquals(lines[0], "TOPIC,HOST,CURRENT-OFFSET,LOG-END-OFFSET,LAG")
	assertStringEquals(lines[1], "topic1,\"consumer-5,consumer-6\",6468718,6468718,0")
	assertStringEquals(lines[2], "topic2,\"consumer-1,consumer-2,consumer-3,consumer-4\",218,218,0")
	assertStringEquals(lines[3], "topic3,consumer-21,26984839,26984839,0")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0] sort[0,2] out..json", fpath)
	actual := unMarshallJsonBytes(execCmd(cmd))
	expected := unMarshallJsonString("[{\"CURRENT-OFFSET\":6468718,\"HOST\":[\"consumer-5\",\"consumer-6\"],\"LAG\":0,\"LOG-END-OFFSET\":6468718,\"TOPIC\":\"topic1\"},{\"CURRENT-OFFSET\":218,\"HOST\":[\"consumer-1\",\"consumer-2\",\"consumer-3\",\"consumer-4\"],\"LAG\":0,\"LOG-END-OFFSET\":218,\"TOPIC\":\"topic2\"},{\"CURRENT-OFFSET\":26984839,\"HOST\":[\"consumer-21\"],\"LAG\":0,\"LOG-END-OFFSET\":26984839,\"TOPIC\":\"topic3\"}]")
	assertDeepEquals(actual, expected)

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0] sort[0,2] head[topic,host,in,out,lag] out..json", fpath)
	actual = unMarshallJsonBytes(execCmd(cmd))
	expected = unMarshallJsonString("[{\"host\":[\"consumer-5\",\"consumer-6\"],\"in\":6468718,\"lag\":0,\"out\":6468718,\"topic\":\"topic1\"},{\"host\":[\"consumer-1\",\"consumer-2\",\"consumer-3\",\"consumer-4\"],\"in\":218,\"lag\":0,\"out\":218,\"topic\":\"topic2\"},{\"host\":[\"consumer-21\"],\"in\":26984839,\"lag\":0,\"out\":26984839,\"topic\":\"topic3\"}]")
	assertDeepEquals(actual, expected)

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[2,1] out..json head[topic,groups,group,stats,in,out,lag]", fpath)
	actual = unMarshallJsonBytes(execCmd(cmd))
	assertIntEquals(len(actual), 3)
	//todo validate the rest

	// 2 group by
	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[0,2]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "TOPIC,HOST,CURRENT-OFFSET,LOG-END-OFFSET,LAG")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[0,2] head[topic,host,in,out,lag]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "topic,host,in,out,lag")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] 'calc(([2]/[3])+1/3)' sort[2]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "TOPIC,HOST,CURRENT-OFFSET,LOG-END-OFFSET,LAG,(CURRENT-OFFSET/LOG-END-OFFSET)+1/3")
	assertStringEquals(lines[1], "topic2,consumer-2,21,21,0,1.33")
	assertStringEquals(lines[17], "topic3,consumer-21,26984839,26984839,0,1.33")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] 'calc(([2]/[3])+1/3)' sort[2] head[topic,host,in,out,lag,calc]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "topic,host,in,out,lag,calc")
	assertStringEquals(lines[1], "topic2,consumer-2,21,21,0,1.33")
	assertStringEquals(lines[17], "topic3,consumer-21,26984839,26984839,0,1.33")

	//todo how to validate the json output

}

func unMarshallJsonBytes(bytes []byte) []map[string]interface{} {
	var array []map[string]interface{}
	json.Unmarshal(bytes, &array)
	return array
}

func unMarshallJsonString(str string) []map[string]interface{} {
	var array []map[string]interface{}
	json.Unmarshal([]byte(str), &array)
	return array
}
