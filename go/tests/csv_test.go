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

	// transform with adding columns and transforming existing cols
	cmd = fmt.Sprintf("cat %v |csv tr..c5..split:/..col[-1]..add tr..c5..split:/..col[0] tr..c5..split:-..col[0,1]..merge:-..add  tr..c5..split:-..col[2:]..merge:-..add out..csv head[topic,partitio,in,out,lag,consumer,host,client,uuid,name,calc] 'calc([0]+\"-\"+[1])'", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "topic,partitio,in,out,lag,consumer,host,client,uuid,name,calc")
	assertStringEquals(lines[1], "topic1,44,808699,808699,0,consumer-5-c6ac0ffe-b453-41ad-a3a4-2b1265735ed3,10.9.27.3,consumer-5,c6ac0ffe-b453-41ad-a3a4-2b1265735ed3,consumer-5,topic1-44")
	assertStringEquals(lines[17], "topic3,0,26984839,26984839,0,consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1,10.9.27.3,consumer-21,2296df7b-b059-4748-9d11-3c6a8a147be1,consumer-21,topic3-0")

	// transform with adding columns and transforming existing cols with sort
	cmd = fmt.Sprintf("cat %v |csv tr..c5..split:/..col[-1]..add tr..c5..split:/..col[0] tr..c5..split:-..col[0,1]..merge:-..add  tr..c5..split:-..col[2:]..merge:-..add out..csv head[topic,partitio,in,out,lag,consumer,host,client,uuid,name,calc] 'calc([0]+\"-\"+[1])' sort[10]:desc", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 19)
	assertStringEquals(lines[0], "topic,partitio,in,out,lag,consumer,host,client,uuid,name,calc")
	assertStringEquals(lines[1], "topic3,0,26984839,26984839,0,consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1,10.9.27.3,consumer-21,2296df7b-b059-4748-9d11-3c6a8a147be1,consumer-21,topic3-0")
	assertStringEquals(lines[17], "topic1,44,808699,808699,0,consumer-5-c6ac0ffe-b453-41ad-a3a4-2b1265735ed3,10.9.27.3,consumer-5,c6ac0ffe-b453-41ad-a3a4-2b1265735ed3,consumer-5,topic1-44")
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

	// 2 group by
	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[0,2]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "TOPIC,HOST,CURRENT-OFFSET,LOG-END-OFFSET,LAG")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1]:count sort[0,2]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "TOPIC,HOST,CURRENT-OFFSET,LOG-END-OFFSET,LAG,count")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0,4")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0,1")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1]:count sort[0,2] head[topic,host,in,out,lag,count]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "topic,host,in,out,lag,count")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0,4")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0,1")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[0,2] head[topic,host,in,out,lag]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "topic,host,in,out,lag")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[0,2] 'calc([2]+[3])'", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "TOPIC,HOST,CURRENT-OFFSET,LOG-END-OFFSET,LAG,CURRENT-OFFSET+LOG-END-OFFSET")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0,6466496")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0,53969678")

	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[0,2] head[topic,host,in,out,lag,calc] 'calc([2]+[3])'", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 9)
	assertStringEquals(lines[0], "topic,host,in,out,lag,calc")
	assertStringEquals(lines[1], "topic1,consumer-6,3233248,3233248,0,6466496")
	assertStringEquals(lines[7], "topic3,consumer-21,26984839,26984839,0,53969678")

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
	assertIntEquals(len(actual), 3) //todo rest

	//fixme the order of the json is getting messed-up due to the last of ordered map
	// at the time of processing the json output
	cmd = fmt.Sprintf("cat %v | csv col[0,6,2,3,4] group[0,1] sort[2,1] out..json..levels:3", fpath)
	var topics []TopicL3
	json.Unmarshal(execCmd(cmd), &topics)

	//todo test the sort & conversion logic when the same column contains mixed (str, int) values

}

func TestCsvHeader(t *testing.T) {
	//dont print header
	fpath := path.Join(getCurrentDir(t), "topics.txt")
	cmd := fmt.Sprintf("cat %v | csv row[17:18] -outhead", fpath)
	lines := execCmdGetLines(cmd)
	assertIntEquals(len(lines), 2)
	assertStringEquals(lines[0], "topic3,0,26984839,26984839,0,consumer-21-2296df7b-b059-4748-9d11-3c6a8a147be1/10.9.27.3,consumer-21")
	assertStringEquals(lines[1], "")

	fpath = path.Join(getCurrentDir(t), "topics_nohead.txt")
	cmd = fmt.Sprintf("cat %v | csv -inhead", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 4)
	assertStringEquals(lines[0], "topic2,1,33,33,0,consumer-1-4d78daa5-13ff-4535-b965-aadc80bcd88e/10.9.27.3,consumer-1")
	assertStringEquals(lines[1], "topic2,2,21,21,0,consumer-2-ebe5eadf-8712-4b84-8951-afc00117e325/10.9.27.3,consumer-2")
	assertStringEquals(lines[2], "topic2,3,30,30,0,consumer-2-ebe5eadf-8712-4b84-8951-afc00117e325/10.9.27.3,consumer-2")
	assertStringEquals(lines[3], "")

	cmd = fmt.Sprintf("cat %v | csv -inhead head[topic,part,in,out,lag,client,consumer]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 5)
	assertStringEquals(lines[0], "topic,part,in,out,lag,client,consumer")
	assertStringEquals(lines[1], "topic2,1,33,33,0,consumer-1-4d78daa5-13ff-4535-b965-aadc80bcd88e/10.9.27.3,consumer-1")
	assertStringEquals(lines[2], "topic2,2,21,21,0,consumer-2-ebe5eadf-8712-4b84-8951-afc00117e325/10.9.27.3,consumer-2")
	assertStringEquals(lines[3], "topic2,3,30,30,0,consumer-2-ebe5eadf-8712-4b84-8951-afc00117e325/10.9.27.3,consumer-2")
	assertStringEquals(lines[4], "")

	//default with no header
	cmd = fmt.Sprintf("cat %v | csv -inhead col[0,2,3,4] group[0]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 2)
	assertStringEquals(lines[0], "topic2,84,84,0")
	assertStringEquals(lines[1], "")

	//explicit headers
	cmd = fmt.Sprintf("cat %v | csv -inhead col[0,2,3,4] group[0] head[topic,in,out,lag]", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 3)
	assertStringEquals(lines[0], "topic,in,out,lag")
	assertStringEquals(lines[1], "topic2,84,84,0")
	assertStringEquals(lines[2], "")

	//incorrect usage, but expected
	cmd = fmt.Sprintf("cat %v | csv -inhead col[0,2,3,4] group[0] head[topic,in,out,lag] -outhead", fpath)
	lines = execCmdGetLines(cmd)
	assertIntEquals(len(lines), 2)
	assertStringEquals(lines[0], "topic2,84,84,0")
	assertStringEquals(lines[1], "")

}

//todo CSV output with complex objects

//todo nohead and -head combinations
//todo json keys with \.

type TopicL3 struct {
	Topic string   `json:"TOPIC"`
	Hosts []HostL3 `json:"HOSTs"`
}
type HostL3 struct {
	Host   string      `json:"HOST"`
	Groups []HostGrpL3 `json:"HOST-group"`
}

type HostGrpL3 struct {
	In  int `json:"LOG-END-OFFSET"`
	Out int `json:"CURRENT-OFFSET"`
	Lag int `json:"LAG"`
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
