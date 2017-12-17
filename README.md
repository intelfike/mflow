# multiple flow
順次処理、通信や入出力の流れを組み合わせて図に表すことができるツールです。

```
独自形式ファイル →[このツール]→ HTML
```

## 変換元ファイル
test.mfw

```
ブラウザ | 表示・入力 | JS | 通信 | PHP

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
```

## 変換後のページ
mflow test.mfw > test.html
