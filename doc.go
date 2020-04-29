package pick

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/liov/pick/utils"
	"github.com/liov/pick/validator"
)

//有swagger,有没有必要做
func GenDoc(modName string, svcs []Service) string {
	buf := new(strings.Builder)
	fmt.Fprintln(buf, "[TOC]")
	fmt.Fprintf(buf, "# %s接口文档  \n", modName)
	fmt.Fprintln(buf, "----------")
	for _, v := range svcs {
		describe, preUrl, _ := v.Service()
		fmt.Fprintf(buf, "# %s  \n", describe)
		fmt.Fprintln(buf, "----------")
		value := reflect.ValueOf(v)
		for j := 0; j < value.NumMethod(); j++ {
			method := value.Type().Method(j)
			if method.Type.NumIn() < 2 || method.Type.NumOut() != 2 {
				continue
			}
			methodInfo := getMethodInfo(value.Method(j))
			if methodInfo.deprecated != nil {
				methodInfo.title += "(废弃)"
			}
			methodInfo.path, methodInfo.version = parseMethodName(method.Name)
			methodInfo.path = preUrl + "/" + methodInfo.path
			methodInfo.path = strings.Replace(methodInfo.path, "${version}", "v"+strconv.Itoa(methodInfo.version), 1)
			//title
			if methodInfo.deprecated != nil {
				fmt.Fprintf(buf, "## ~~%s-v%d(废弃)(`%s`)~~  \n", methodInfo.title, methodInfo.version, methodInfo.path)
			} else {
				fmt.Fprintf(buf, "## %s-v%d(`%s`)  \n", methodInfo.title, methodInfo.version, methodInfo.path)
			}
			//api
			fmt.Fprintf(buf, "**%s** `%s` _(Principal %s)_  \n", methodInfo.method, methodInfo.path, methodInfo.getPrincipal())

			fmt.Fprint(buf, "### 接口记录  \n")
			fmt.Fprint(buf, "|版本|操作|时间|负责人|日志|  \n")
			fmt.Fprint(buf, "| :----: | :----: | :----: | :----: | :----: |  \n")
			fmt.Fprintf(buf, "|%s|%s|%s|%s|%s|  \n", methodInfo.createlog.version, "创建", methodInfo.createlog.date, methodInfo.createlog.auth, methodInfo.createlog.log)
			if len(methodInfo.changelog) != 0 || methodInfo.deprecated != nil {
				for _, clog := range methodInfo.changelog {
					fmt.Fprintf(buf, "|%s|%s|%s|%s|%s|  \n", clog.version, "变更", clog.date, clog.auth, clog.log)
				}
				if methodInfo.deprecated != nil {
					fmt.Fprintf(buf, "|%s|%s|%s|%s|%s|  \n", methodInfo.deprecated.version, "删除", methodInfo.deprecated.date, methodInfo.deprecated.auth, methodInfo.deprecated.log)
				}
			}

			fmt.Fprint(buf, "### 参数信息  \n")
			if method.Type.NumIn() == 3 {
				fmt.Fprint(buf, "|字段名称|字段类型|字段描述|校验要求|  \n")
				fmt.Fprint(buf, "| :----  | :----: | :----: | :----: |  \n")
				params := getParamTable(method.Type.In(2).Elem(), "")
				for i := range params {
					fmt.Fprintf(buf, "|%s|%s|%s|%s|  \n", params[i].json, params[i].typ, params[i].annotation, params[i].validator)
				}

			} else {
				fmt.Fprint(buf, "无需参数")
			}
			fmt.Fprint(buf, "__请求示例__  \n")
			fmt.Fprint(buf, "```json  \n")
			newParam := reflect.New(method.Type.In(2).Elem()).Interface()
			utils.Mock(newParam)
			data, _ := json.MarshalIndent(newParam, "", "\t")
			fmt.Fprint(buf, string(data), "  \n")
			fmt.Fprint(buf, "```  \n")
			fmt.Fprint(buf, "### 返回信息  \n")
			fmt.Fprint(buf, "|字段名称|字段类型|字段描述|  \n")
			fmt.Fprint(buf, "| :----  | :----: | :----: | \n")
			params := getParamTable(method.Type.Out(0).Elem(), "")
			for i := range params {
				fmt.Fprintf(buf, "|%s|%s|%s|  \n", params[i].json, params[i].typ, params[i].annotation)
			}
			fmt.Fprint(buf, "__返回示例__  \n")
			fmt.Fprint(buf, "```json  \n")
			newRes := reflect.New(method.Type.Out(0).Elem()).Interface()
			utils.Mock(newRes)
			data, _ = json.MarshalIndent(newRes, "", "\t")
			fmt.Fprint(buf, string(data), "  \n")
			fmt.Fprint(buf, "```  \n")
		}
	}
	return buf.String()
}

type ParamTable struct {
	json, annotation, typ, validator string
}

func getParamTable(param reflect.Type, pre string) []*ParamTable {
	param = utils.OriginalType(param)
	newParam := reflect.New(param).Interface()
	var res []*ParamTable
	for i := 0; i < param.NumField(); i++ {
		/*		if param.AssignableTo(reflect.TypeOf(response.File{})) {
				return "下载文件"
			}*/
		var p ParamTable
		field := param.Field(i)
		if field.Anonymous {
			continue
		}
		json := strings.Split(field.Tag.Get("json"), ",")[0]
		if json == "-" {
			continue
		}
		if json == "" {
			p.json = pre + utils.ConvertToCamelCase(json)
		} else {
			p.json = pre + json
		}
		p.annotation = field.Tag.Get("annotation")
		if p.annotation == "-" {
			p.annotation = p.json
		}
		p.typ = getJsType(field.Type)
		if valid := validator.Trans(validator.Validate.StructPartial(newParam, field.Name)); valid != "" {
			p.validator = valid[len(p.annotation):]
		}
		if p.typ == "object" || p.typ == "[]object" {
			p.json = "**" + p.json + "**"
			res = append(res, &p)
			sub := getParamTable(field.Type, json+".")
			res = append(res, sub...)
		} else {
			res = append(res, &p)
		}
	}
	return res
}

func getJsType(typ reflect.Type) string {
	t := time.Time{}
	if typ == reflect.TypeOf(t) || typ == reflect.TypeOf(&t) {
		return "date"
	}
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Array, reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return "string"
		}
		return "[]" + getJsType(typ.Elem())
	case reflect.Ptr:
		return getJsType(typ.Elem())
	case reflect.Struct:
		return "object"
	case reflect.Bool:
		return "boolean"
	}
	return "string"
}
