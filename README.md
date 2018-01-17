# multiple flow
並列処理、通信や入出力の流れを組み合わせて図に表すことができるツールです。

```
# 独自形式ファイル → HTML
mflow test.mfw
```
test.htmlが出力されます。

# 使用例
test.mfw

```
テスト用の表です
ここにはページの詳細を書くことができます。

=== フロー名 === [ブラウザ] 表示・入力 [JS] 通信 [PHP]

ブラウザのテキストエリアのデータをPOSTするプログラムのフロー
ここにはフローの詳細を書くことができます。

---

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

```

## 変換後のページ
<img src="https://github.com/intelfike/mflow/blob/master/sample_image/screen_shot.png">
