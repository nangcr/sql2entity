package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Column 储存列信息
type Column struct {
	Name         string
	Type         string
	Nullable     bool
	Length       int
	IsPrimaryKey bool
	Comment      string
}

// Table 储存表结构信息
type Table struct {
	TableName string
	Columns   []Column
}

// FormatName 删除多余的下划线并大写首字母
func FormatName(str string) (result string) {
	s := strings.Split(str, "_")
	for _, v := range s {
		result += strings.ToUpper(v[0:1]) + v[1:]
	}
	return
}

// MakeColumn 从字符串读取列定义数据并返回列结构
func MakeColumn(columns string) (c Column, ok bool) {
	//使用空格分词
	values := strings.Split(strings.TrimSpace(columns), " ")

	//判断是否需要处理
	if !strings.Contains(values[0], "`") {
		ok = false
		return
	}
	//提取列名称
	c.Name = strings.Split(values[0], "`")[1]

	//提取NOT NULL属性
	if strings.Contains(columns, "NOT NULL") {
		c.Nullable = false
	} else {
		c.Nullable = true
	}

	//提取列类型
	if strings.Contains(values[1], "int") {
		c.Type = "int"
	}
	if strings.Contains(values[1], "varchar") {
		c.Type = "string"
		if strings.Contains(values[1], "(") {
			_, err := fmt.Sscanf(values[1], "varchar(%d)", &c.Length)
			if err != nil {
				_, _ = fmt.Fprint(os.Stderr, err.Error())
			}
		}
	}
	if strings.Contains(values[1], "date") {
		if c.Nullable {
			c.Type = "DateTime?"
		} else {
			c.Type = "DateTime"
		}
	}
	if strings.Contains(values[1], "decimal") {
		c.Type = "decimal"
	}
	if strings.Contains(values[1], "double") {
		c.Type = "double"
	}

	//提取注释信息
	if strings.Contains(columns, "'") {
		c.Comment = strings.Split(columns, "'")[1]
	}

	ok = true
	return
}

// MakeTable 从字节流中读取数据并返回Table结构
func MakeTable(b []byte) (table Table, err error) {
	str := string(b)

	//抛弃无用数据
	str = strings.Split(str, "CREATE TABLE ")[1]

	//提取表名
	table.TableName = strings.Split(str, "`")[1]

	columns := strings.Split(str, "\n")[1:]
	for _, c := range columns {
		temp, ok := MakeColumn(c)
		if ok {
			table.Columns = append(table.Columns, temp)
		}

		//标注主键
		if strings.Contains(c, "PRIMARY KEY") {
			for k, tc := range table.Columns {
				if strings.Contains(c, tc.Name) {
					table.Columns[k].IsPrimaryKey = true
				}
			}
		}
	}
	return
}

// GenCode 根据Table对象生成目标代码
func (table Table) GenCode(entityName string) (result string, err error) {
	const (
		Comment      = "/// <summary>\n/// %s\n/// </summary>\n"
		TableName    = "[Table(\"%s\")]\n"
		Key          = "[Key]\n"
		Column       = "[Column(\"%s\")]\n"
		StringLength = "[StringLength(%d)]\n"
		Required     = "[Required]\n"
		Class        = "public class %s : IEntity\n"
		Function     = "public %s %s { get; set; }\n"
	)

	//表名
	result += fmt.Sprintf(Comment, "")
	result += fmt.Sprintf(TableName, table.TableName)
	//声明类
	result += fmt.Sprintf(Class, entityName)
	result += fmt.Sprintf("{\n")

	for i, c := range table.Columns {
		if i != 0 {
			result += fmt.Sprintf("\n")
		}
		//注释
		result += fmt.Sprintf(Comment, c.Comment)
		//主键
		if c.IsPrimaryKey {
			result += fmt.Sprintf(Key)
		}
		//NOT NULL
		if !c.Nullable {
			result += fmt.Sprintf(Required)
		}
		//列名
		result += fmt.Sprintf(Column, c.Name)
		//长度
		if c.Type == "string" {
			result += fmt.Sprintf(StringLength, c.Length)
		}
		//属性
		result += fmt.Sprintf(Function, c.Type, FormatName(c.Name))
	}

	result += fmt.Sprintf("}\n")
	return
}

func main() {
	entityName := ""
	inputFileName := ""
	//判断命令行参数是否合法
	if len(os.Args) == 2 {
		inputFileName = os.Args[1]
	} else if len(os.Args) == 3 {
		entityName = os.Args[1]
		inputFileName = os.Args[2]
	} else {
		_, _ = fmt.Fprint(os.Stderr, "Usage: sql2entity [entity name] <file>")
		return
	}

	//读取配置文件
	prefix, err := ioutil.ReadFile("prefix")
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
	}
	suffix, err := ioutil.ReadFile("suffix")
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
	}

	//读取需要操作的文件
	b, err := ioutil.ReadFile(inputFileName)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
		return
	}

	fmt.Println("Read file succeed")

	//建立模型
	table, err := MakeTable(b)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
		return
	}
	fmt.Println("Model build succeed")

	if entityName == "" {
		entityName = table.TableName
	}
	fmt.Println("Entity class name:", entityName)

	//生成代码
	code, err := table.GenCode(entityName)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
		return
	}
	fmt.Println("Generate code succeed")

	err = ioutil.WriteFile(entityName+".cs",
		[]byte(fmt.Sprintln(string(prefix))+
			fmt.Sprint(code)+
			fmt.Sprint(string(suffix))),
		os.ModeAppend)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
		return
	}
	fmt.Println("Write file succeed")
}
