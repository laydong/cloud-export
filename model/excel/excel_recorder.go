package excel

import (
	"cloud-export/model/helper"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tidwall/gjson"
)

type ExcelRecorder struct {
	FilePath         string
	ExcelFp          *excelize.File
	Sheet            string
	Limit            int
	Page             int
	FileNameTemplate string
}

func NewExcelRecorder(path string) *ExcelRecorder {
	instance := &ExcelRecorder{
		FilePath: path,
		ExcelFp:  excelize.NewFile(),
		Sheet:    "Sheet1",
	}
	return instance
}

func NewExcelRecorderPage(template string, limit int) *ExcelRecorder {
	ins := &ExcelRecorder{
		ExcelFp:          excelize.NewFile(),
		Sheet:            "Sheet1",
		Page:             1,
		Limit:            limit,
		FileNameTemplate: template,
	}
	return ins
}

// JsonListWrite 将json列表写入excel中
func (e *ExcelRecorder) JsonListWrite(p Pos, jsonStr string, isFirst bool) Pos {
	result := gjson.Parse(jsonStr)
	keys := make([]string, 0)
	for i, line := range result.Array() {
		if i == 0 {
			keys = JsonKeys(line.String())
			if !isFirst {
				// 不写首行数据
				continue
			}
		}
		p = e.JsonWrite(p, keys, line.String())
	}
	return p
}

// JsonWrite 单条json 按照keys的键顺序写入excel
// ajson格式 `{"key1":"val1", "key2":"val2"}`
func (e *ExcelRecorder) JsonWrite(p Pos, keys []string, ajson ...string) Pos {
	for _, line := range ajson {
		aline := gjson.Parse(line).Value().(map[string]interface{})
		lineValues := helper.Map2Arr(aline, keys)

		// 写入一行
		// e.ExcelFp.SetSheetRow(e.Sheet, p.String(), &lineValues)

		// 按单元格依次写入，并支持单元格样式，例如：
		// {"text":"文本","font":{"color":"#ed0000"}}
		// {"text":"文本","fill":{"color":"#ed0000"}}
		// {"text":"文本","fill":{"color":"#ed0000"},"font":{"color":"#ffffff"}}
		for i, value := range lineValues {
			p.X = i + 1
			position := p.String()
			str, ok := value.(string)
			if ok {
				cellValue := ParseCellValue(e.ExcelFp, str)
				if cellValue.Style != 0 {
					e.ExcelFp.SetCellStyle(e.Sheet, position, position, cellValue.Style)
				}
				e.ExcelFp.SetCellValue(e.Sheet, position, cellValue.Text)
			} else {
				e.ExcelFp.SetCellValue(e.Sheet, position, value)
			}
		}
		p.Y += 1
	}
	return p
}

// Save 保存
func (e *ExcelRecorder) Save() error {
	if e.FileNameTemplate != "" {
		e.FilePath = fmt.Sprintf(e.FileNameTemplate, e.Page)
	}
	helper.TouchDir(e.FilePath)
	if err := e.ExcelFp.SaveAs(e.FilePath); err != nil {
		return err
	}
	return nil
}

func List2Arrs(lines []map[string]interface{}, keys []string) [][]interface{} {
	ret := make([][]interface{}, 0, len(lines))
	for _, line := range lines {
		tmp := helper.Map2Arr(line, keys)
		ret = append(ret, tmp)
	}
	return ret
}

// 分页写入
func (e *ExcelRecorder) WritePagpenate(p Pos, lineJson string, htable string, isFirst bool) Pos {
	lines := gjson.Parse(lineJson).Array()
	if htable == "" {
		htable = lines[0].String()
		lines = lines[1:]
	}
	keys := JsonKeys(htable)
	// 第一次调用 需要写表头
	if isFirst {
		p = e.JsonWrite(p, keys, htable)
	}
	for _, aline := range lines {
		if p.Y >= e.Limit {
			// 保存文件
			e.Save()
			// 新生成文件
			e.ExcelFp = excelize.NewFile()
			e.Page += 1
			p.Y = 1
			// 写表头
			p = e.JsonWrite(p, keys, htable)
		}
		p = e.JsonWrite(p, keys, aline.String())
	}
	return p
}

type Pos struct {
	X, Y           int
	Row, Col, Addr string
}

// Convert 坐标转Excel地址
func (p *Pos) Convert() {
	p.Col = x2col(p.X)
	p.Row = strconv.Itoa(p.Y)
	p.Addr = p.Col + p.Row
}

func (p *Pos) String() string {
	// 先转化
	p.Convert()
	return p.Addr
}

func x2col(x int) string {
	result := ""
	for x > 0 {
		x--
		result = string(rune(x%26+'A')) + result
		x = x / 26
	}
	return result
}

func JsonKeys(json string) []string {
	keys := make([]string, 0)
	gjson.Parse(json).ForEach(func(key, value gjson.Result) bool {
		keys = append(keys, key.String())
		return true
	})
	return keys
}

// CellValue 解析后，带样式的单元格
type CellValue struct {
	Style int
	Text  string
}

// CellValueJson 解析前，单元格的 JSON 内容
type CellValueJson struct {
	Text string `json:"text"`
	Fill struct {
		Color string `json:"color"`
	} `json:"fill"`
	Font struct {
		Color string `json:"color"`
	} `json:"font"`
}

// ParseCellValue 解析单元格的样式和文本内容
func ParseCellValue(file *excelize.File, value string) CellValue {
	cellValue := CellValue{
		Text: value,
	}
	if len(value) >= 2 && value[0:1] == "{" {
		data := &CellValueJson{}
		if err := json.Unmarshal([]byte(value), data); err == nil {
			cellValue.Text = data.Text
			if data.Fill.Color != "" || data.Font.Color != "" {
				styleData := map[string]interface{}{}
				if data.Fill.Color != "" {
					styleData["fill"] = map[string]interface{}{
						"type":    "pattern",
						"color":   []string{data.Fill.Color},
						"pattern": 1,
					}
				}
				if data.Font.Color != "" {
					styleData["font"] = map[string]interface{}{
						"color": data.Font.Color,
					}
				}
				styleJson, jsonErr := json.Marshal(styleData)
				if jsonErr != nil {
					return cellValue
				}
				style, styleErr := file.NewStyle(string(styleJson))
				if styleErr != nil {
					return cellValue
				}
				cellValue.Style = style
			}
		}
	}
	return cellValue
}
