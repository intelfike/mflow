package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Column struct {
	ID       int
	cellType string // work OR arrow
	title    string
}

func parseColumns(headers []string) map[string]*Column {
	columns := map[string]*Column{}
	for n, v := range headers {
		col := new(Column)
		col.ID = n
		col.title = v
		col.cellType = "work"
		if strings.HasPrefix(v, "(") && strings.HasSuffix(v, ")") {
			col.cellType = "arrow"
		}
		columns[v] = col
	}
	return columns
}

type Cell struct {
	ID       int
	Type     string
	ColumnID int
	text     string
	origin   string
}

func main() {
	if len(os.Args) != 2 {
		printUsage()
		return
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	lines := strings.Split(string(b), "\n")

	// ヘッダの作成
	headers, err := fetchHeader(lines)
	if err != nil {
		fmt.Println(err)
		return
	}
	cols := parseColumns(headers)
	data := make([][]string, 0)
	wid := 0
	row := 0
	curTag := headers[0]
	changed := false

	var description string
	for n, line := range lines[1:] {
		if line == "---" {
			description = strings.Join(lines[n+2:], "\n")
			break
		}
		if line == "" {
			continue
		}
		curID := cols[curTag].ID
		// タグを切り分ける
		destTag, text, err := splitTag(line)
		if err != nil { // ワークなら
			if !changed {
				// もし前と同じカラムなら次の行を追加
				data = append(data, make([]string, len(headers)))
				row++
			}
			wid++
			// 書き込み
			data[row-1][curID] = strconv.Itoa(wid) + ", " + line
			changed = false
			continue
		}
		// アローなら
		_, ok := cols[destTag]
		if !ok {
			fmt.Println("Error:", n+1, "行目、", line, "\nカラムが見つかりませんでした:", destTag)
			os.Exit(1)
		}
		destID := cols[destTag].ID
		arrow := "=>"
		if curID > destID {
			arrow = "<="
		}
		arrowID := (curID + destID) >> 1

		if len(data) == 0 || data[row-1][arrowID] != "" {
			// アロー書き込み先に文字があれば行を新しく作る
			data = append(data, make([]string, len(headers)))
			row++
		}
		if len(data) > 1 && data[row-2][arrowID] != "" {
			// 直上の要素が空じゃなければ一個隙間を開ける
			data = append(data, make([]string, len(headers)))
			row++
		}
		// 値を登録
		if text != "" {
			detail := detailReg.FindString(text)
			text = strings.TrimSuffix(text, detail)
			data[row-1][arrowID] = arrow + "[" + text + "]" + arrow + detail
		} else {
			data[row-1][arrowID] = arrow
		}

		// 後置処理
		curTag = destTag
		changed = true
	}
	ht := createTable(headers, data)
	fmt.Println(`<!DOCTYPE html>
<title>` + os.Args[1] + `</title>
<mate charset="utf-8">
<style>
body {
	font-size: 18px;
}
table {
	border-spacing: 0;
	border: 1px solid black;
}
th{
	border-bottom: 1px solid black;
}
th, td{
	margin: 0;
	border-right: 1px solid black;
	vertical-align: top;
}
.col-1, .col-3, .col-5, .col-7, .col-9, .col-11{
	background-color: #F8F8F8;
	text-align: center;
	font-size: 80%;
}

</style>`)
	fmt.Println(ht)
	fmt.Println(description)
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

var detailReg = regexp.MustCompile("\\([^()]+\\)\\s*$")

// HTMLを作成する
func createTable(headers []string, data [][]string) string {
	ht := "<table>\n"
	ht += "<tr>\n"
	for n, v := range headers {
		ht += `	<th class="col-` + strconv.Itoa(n) + `">`
		ht += v
		ht += "</th>\n"
	}
	ht += "</tr>\n"

	wid := 0
	for n, v := range data {
		wid++
		swid := strconv.Itoa(wid)
		ht += "<tr>\n"
		for n2, v2 := range v {
			ht += `	<td id="wflow-` + swid + `" class="col-` + strconv.Itoa(n2) + " row-" + strconv.Itoa(n) + `">`
			detail := detailReg.FindString(v2)
			if detail != "" {
				ht += `<a href="#" onclick="alert('` + strings.Trim(detail, " 　\t()") + `')">`
				ht += strings.TrimSuffix(v2, detail)
				ht += `</a>`
			} else {
				ht += v2
			}
			ht += "</td>\n"
		}
		ht += "</tr>\n"
	}
	ht += "</table>\n"
	return ht
}

func fetchHeader(lines []string) ([]string, error) {
	head := lines[0]
	headers := strings.Split(head, "|")
	for n, v := range headers {
		headers[n] = strings.Trim(v, " 　\t")
	}
	return headers, nil
}

func printUsage() {
	fmt.Println(`#### 使い方:
mflow [filename]


#### mflowファイルの書き方:
## 一行目 カラムの定義
カラム1 | (中継1) | カラム2

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
↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
ブラウザ | (表示・入力) | JS | (通信) | PHP

[ブラウザ]
ページ訪問
フォーム入力
submitボタンクリック

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
