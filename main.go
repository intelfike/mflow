package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Column struct {
	ID       int
	cellType string // work OR arrow
	title    string
	cells    []*Cell
}

func parseColumns(headers []string) ([]*Column, map[string]*Column) {
	colMap := map[string]*Column{}
	colArray := make([]*Column, len(headers))
	for n, v := range headers {
		col := new(Column)
		col.ID = n
		col.cellType = "arrow"
		if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
			col.cellType = "work"
		}
		v = strings.Trim(v, " 　\t[]")
		col.title = v
		colMap[v] = col
		colArray[n] = col
	}
	return colArray, colMap
}
func fetchHeader(lines []string) ([]string, error) {
	head := lines[0]
	headers := regexp.MustCompile("\\[[^\\]]+\\]|[^\\[]*").FindAllString(head, -1)
	return headers, nil
}

type Cell struct {
	ID             int
	Type           string
	destTag        string
	prefix, suffix string
	text, origin   string
	detail         string
	bgcolor        string
}

var detailReg = regexp.MustCompile("\\([^()]+\\)\\s*$")

func parseCell(text string) *Cell {
	cell := new(Cell)
	cell.origin = text
	destTag, text, err := splitTag(text)
	if err != nil {
		cell.Type = "work"
	} else {
		cell.destTag = destTag
		cell.Type = "arrow"
	}

	detail := detailReg.FindString(text)
	if detail != "" {
		text = strings.TrimSuffix(text, detail)
		cell.detail = strings.Trim(detail, " 　\t()")
	}

	if strings.HasPrefix(text, "#") {
		text = strings.TrimPrefix(text, "#")
		cell.bgcolor = "yellow"
	}

	text = strings.Trim(text, " 　\t")

	cell.text = text
	return cell
}

func main() {
	if len(os.Args) != 2 {
		printUsage()
		return
	}
	if !strings.HasSuffix(os.Args[1], ".mfw") {
		fmt.Println("読み込むファイルの拡張子は.mfwで統一してください")
		return
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	os.Args[1] = strings.TrimSuffix(os.Args[1], ".mfw")
	lines := strings.Split(string(b), "\n")

	colArray, data, nextLine := createCells(lines)
	ht := createTable(colArray, data)
	_, file := filepath.Split(os.Args[1])
	result := `<!DOCTYPE html>
<title>` + file + `</title>
<mate charset="utf-8">
<style>
body {
	background-color: #EEEEEE;
	font-size: 18px;
}
table {
	border-spacing: 0;
	margin: 16px;
	border-collapse: collapse;
}
th{
	background-color: white !important;
	padding: 8px;
	border: 1px solid black;
}
th, td{
	margin: 0;
	vertical-align: top;
}
.work{
	border-right: 1px solid black;
	border-left: 1px solid black;
	background-color: white;
}
.arrow{
	background-color: rgba(0,0,0,0);
	text-align: center;
	font-size: 80%;
}
.left.work{
	background-color: #CCC;
}
.left.arrow{
	background-color: #AAA;
}
#tip{
	display: none;
	position: absolute;
	background-color: #DEF;
	border: 1px solid black;
	padding: 8px;
	pointer-events: none;
}

</style>

<script>
function showTip(e, text){
	tip.innerHTML = text
	tip.style.left = (e.clientX+1) + 'px'
	tip.style.top = (e.clientY+1) + 'px'
	tip.style.display = 'block'
}
function hideTip(){
	tip.style.display = 'none'
}
</script>

<div id="tip"></div>

<h1>` + file + `</h1>`
	result += ht
	for n, line := range lines[nextLine:] {
		if line == "---" {
			result += strings.Join(lines[nextLine+n+1:], "\n")
			break
		}
	}
	ioutil.WriteFile(os.Args[1]+".html", []byte(result), 0777)
}

func createCells(lines []string) ([]*Column, [][]*Cell, int) {
	// ヘッダの作成
	headers, err := fetchHeader(lines)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	colArray, cols := parseColumns(headers)
	data := make([][]*Cell, 0)
	wid := 0
	row := 0
	curTag := strings.Trim(headers[0], " 　\t[]")
	changed := true
	nextLine := 0

	for n, line := range lines[1:] {
		nextLine = n
		if line == "---" {
			break
		}
		if line == "" {
			continue
		}
		curID := cols[curTag].ID
		// タグを切り分ける
		cell := parseCell(line)
		if cell.Type == "work" { // ワークなら
			if !changed {
				// もし前と同じカラムなら次の行を追加
				data = append(data, make([]*Cell, len(headers)))
				row++
			}
			wid++
			// 書き込み
			cell.ID = wid
			cell.prefix = strconv.Itoa(wid) + ", "
			data[row-1][curID] = cell
			changed = false
			continue
		}
		// アローなら
		_, ok := cols[cell.destTag]
		if !ok {
			fmt.Println("Error:", n+1, "行目、", line, "\nカラムが見つかりませんでした:", cell.destTag)
			os.Exit(1)
		}
		destID := cols[cell.destTag].ID
		cell.prefix = "==["
		cell.suffix = "]=>"
		dummyArrow := "==>"
		leftID := curID
		rightID := destID
		if curID > destID {
			cell.prefix = "<=["
			cell.suffix = "]=="
			dummyArrow = "<=="
			leftID, rightID = rightID, leftID
		}
		if cell.text == "" {
			cell.prefix = ""
			cell.suffix = ""
			cell.text = dummyArrow
		}

		arrowID := (leftID + rightID) >> 1
		if len(data) == 0 || data[row-1][arrowID] != nil {
			// アロー書き込み先に文字があれば行を新しく作る
			newRow := make([]*Cell, len(headers))
			newRow[curID] = &Cell{text: "↓", Type: "dummy_work"}
			data = append(data, newRow)
			if row >= 1 {
				data[row][curID], data[row-1][curID] = data[row-1][curID], data[row][curID]
			}
			row++
		}
		if len(data) > 1 && data[row-2][arrowID] != nil {
			// 直上の要素が空じゃなければ一個隙間を開ける
			newRow := make([]*Cell, len(headers))
			newRow[curID] = &Cell{text: "↓", Type: "dummy_work"}
			data = append(data, newRow)
			if row >= 1 {
				data[row][curID], data[row-1][curID] = data[row-1][curID], data[row][curID]
			}
			row++
		}
		if leftID == rightID {
			continue
		}
		// 値を登録
		for n := leftID + 1; n < rightID; n++ {
			data[row-1][n] = &Cell{text: dummyArrow, Type: "dummy_arrow"}
		}
		data[row-1][arrowID] = cell

		// 後置処理
		curTag = cell.destTag
		changed = true
	}
	return colArray, data, nextLine
}

var tagReg = regexp.MustCompile("^\\[[^\\]]+\\]")

// アローかどうか判定する
// 戻り値: 宛先カラム、内容、エラー
func splitTag(text string) (string, string, error) {
	col := tagReg.FindString(text)
	text = strings.TrimPrefix(text, col)
	col = strings.Trim(col, "[]")
	if col == "" {
		return "", text, errors.New("Not tag.")
	}
	return col, text, nil
}

// HTMLを作成する
func createTable(headers []*Column, data [][]*Cell) string {
	ht := "<table>\n"
	ht += "<tr>\n"
	for n, v := range headers {
		ht += `	<th class="col-` + strconv.Itoa(n) + `">`
		ht += v.title
		ht += "</th>\n"
	}
	ht += "</tr>\n"

	wid := 0
	for n, v := range data {
		wid++
		swid := strconv.Itoa(wid)
		ht += "<tr>\n"
		leftFlag := true
		for n2, v2 := range v {
			if v2 != nil {
				leftFlag = false
				events := ""
				class := ""
				style := ""
				if v2.detail != "" {
					events = ` onmouseover="showTip(event, '` + v2.detail + `')" onmouseout="hideTip()"`
				}
				switch v2.Type {
				case "work", "dummy_work":
					class = "work"
					style = `background-color:` + v2.bgcolor + `;`
				case "arrow", "dummy_arrow":
					class = "arrow"
					if n2%2 == 0 {
						style = `border-top:1px solid black; border-bottom:1px solid black;`
					}
				default:
					fmt.Println(v2.Type)
				}
				ht += `	<td class="` + class + ` col-` + strconv.Itoa(n2) + " row-" + strconv.Itoa(n) + `" style="` + style + `"` + events + `>`

				if v2.detail != "" {
					ht += `<a href="#" ` + events + `>`
					ht += v2.prefix + v2.text + v2.suffix
					ht += `</a>`
				} else {
					ht += v2.prefix + v2.text + v2.suffix
				}
			} else {
				class := ""
				if n2%2 == 0 {
					class = "work "
				} else {
					class = "arrow "
				}
				if leftFlag {
					class += "left "
				}
				ht += `	<td id="wflow-` + swid + `" class="empty ` + class + `col-` + strconv.Itoa(n2) + " row-" + strconv.Itoa(n) + `">`
			}
			ht += "</td>\n"
		}
		ht += "</tr>\n"
	}
	ht += "</table>\n"
	return ht
}
func printUsage() {
	fmt.Println(`#### 使い方:
mflow [filename]


#### mflowファイルの書き方:
## 一行目 カラムの定義
[カラム1] (中継1) [カラム2]

## 二行目以降 セルの定義
# 行ごとにセル(ワーク・アロー)を定義します。

## ワーク
# ()を利用すると、内容の文字をクリックで詳細を見ることがでるようになります。
内容
内容(詳細)

## アロー
# カラムを切り替えることができます。
# タグは、カラム名を入れてください。矢印の宛先のカラム名になります。
[タグ]
[タグ]説明
[タグ]説明(詳細)

#### 例:
test.mfw
↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
[ブラウザ] 表示・入力 [JS] 通信 [PHP]

[ブラウザ]
ページ訪問
フォーム入力
#submitボタンクリック

[JS]POST (http://example.com/Api/?[フォームデータ])
フォームデータをPOST

[PHP]
フォームのデータを読みだして保存
メッセージを返信(メッセージ: 投稿が完了しました)

[JS]JSON
受信したJSONからメッセージを取り出す
HTMLで表示

[ブラウザ]
メッセージを閉じる
終了


---
テキストエリアのデータをPOSTするプログラムのフロー<br>
↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑
`)
}
