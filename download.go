package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var resource_path string
var num int = 0

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	resource_path = GetCurrentDirectory()

	fmt.Println("下载资源所在目录：" + resource_path)

	fmt.Println("正在执行....")

	//读取json
	json_str, err1 := ioutil.ReadFile(resource_path + "/data.json")
	if err1 != nil {
		fmt.Println(err1)
		os.Exit(1)
	}

	//将json数据转换成的map
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(json_str), &m)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	project_name, ok := m["project_name"].(string)

	if !ok || project_name == "" || !CreateFolder(project_name) {
		fmt.Println("project_name error")
		os.Exit(3)
	}
	start_time := time.Now().Unix()
	//开启线程数
	thread_num := 4
	ch := make(chan bool, thread_num)

	if project_data, ok := m["data"].([]interface{}); ok {
		length := len(project_data)

		fmt.Println("总任务数：" + fmt.Sprintf("%d", length))

		//每组用户数
		group_size := length / thread_num
		//不够多协程分配则改为单协程
		if group_size == 0 {
			thread_num = 1
		}
		var pd []interface{}
		for i := 0; i < thread_num; i++ {
			//分割每组数据
			start_index := i * int(group_size)
			end_index := (i + 1) * int(group_size)
			if i == thread_num-1 {
				pd = project_data[start_index:]
			} else {
				pd = project_data[start_index:end_index]
			}
			//fmt.Println(len(pd))
			//执行协程
			go SaveUserPics(pd, project_name, ch)
		}
	} else {
		fmt.Println("project_data error")
		os.Exit(4)
	}
	for i := 0; i < thread_num; i++ {
		<-ch
	}
	use_time := time.Now().Unix() - start_time
	fmt.Println("下载结束,用时" + strconv.Itoa(int(use_time)) + "秒！！！10秒后自动关闭！！！")
	time.Sleep(time.Duration(10) * time.Second)
}

/**
*下载保存
 */
func SaveUserPics(project_data []interface{}, project_name string, ch chan bool) {
	for _, userItem := range project_data {
		userMap := userItem.(map[string]interface{})
		user_name, ok := userMap["user_name"].(string)
		if ok && user_name != "" && CreateFolder(project_name+"/"+user_name) {
			if user_data, ok := userMap["data"].([]interface{}); ok {
				for _, infoItem := range user_data {
					infoMap := infoItem.(map[string]interface{})
					name := infoMap["name"].(string)
					pic := infoMap["pic"].(string)
					resp, _ := http.Get(pic)
					body, _ := ioutil.ReadAll(resp.Body)
					out, _ := os.Create(resource_path + "/" + project_name + "/" + user_name + "/" + name + path.Ext(pic))
					io.Copy(out, bytes.NewReader(body))
				}
			}
		}
		num++
		fmt.Println("已完成：" + fmt.Sprintf("%d", num) + "    用户名：" + user_name)
	}
	ch <- true
	// 写文件记录一下上一次下载到第几个
	// var d1 = []byte(fmt.Sprintf("%d", num))
	// ioutil.WriteFile("./output2.txt", d1, 0666)
}

/**
 * 判断文件夹是否存在,并创建
 */
func CreateFolder(path string) bool {
	if path == "" {
		return false
	}
	path = resource_path + "/" + path
	//	fmt.Println(path)
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return false
		} else {
			return true
		}
	}
	return false
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}
