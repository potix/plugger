# plugger

## 説明
go の buildmode で c-shared でビルドした shared library (.so) を別の go から 動的に呼び出せるようにするフレームワーク。
実装は割と力技。

## plugger
plugin を利用する側が import するライブラリ。 NewPlugger を実行してインスタンスを作成する。

### Pluggerインスタンス

- Load : 与えられたパス以下の shared library を探してロードする。
- Unload : 指定された名前を持つライブラリをアンロードする。
- GetBuildVersion : ビルドバージョンを取得する。
- GetPluginNames : ロードしたプラグインの名前一覧を取得する。
- ExistsPluginNames : プラグイン名の一覧を渡して、存在するものの一覧が返る。
- NewPlugin  : プラグインハンドルインスタンスの作成。plugin呼び出し後の返却される結果を格納する構造体を生成する関数をセットする必要がある。
- FreePlugin : プラグインハンドルインスタンスの削除。
- Free       : Pluggerインスタンスが持つ全てのリソースを解放する。

### PluginHandleインスタンス
後術するPluginインターフェイスと対応している

- Init    : プラグインの初期化をする。
- Start   : プラグインを開始する。
- Stop    : プラグインを停止する。
- Reload  : プラグインを再設定をする。
- Fini    : プラグインの終了処理をする。
- Command : プラグインにコマンドを送る。
- EventOn : プラグインからのイベント受けて処理結果を返すハンドラを登録する。

## plugin
plugin 側の実装が import するライブラリ。 Plugin インターフェイスを満たしたインスタンスを用意する必要がある。
pluginは shared library なので main 関数には何も記述しないで、 -buildmode=c-shared でコンパイルする必要がある。
init()関数内で用意された特定の初期化関数を呼び出す必要がある。
EventEmit関数を使ってplugin側からイベントを送ることができる

### 初期化関数

- SetPluginName          : プラグインの名前をセットする。呼び出し側はこの名前を利用してプラグインハンドラを生成する。
- SetNewPluginFunc       : プラグインインスタンスを生成する関数をセットする。
- SetNewConfigFunc       : プラグインに渡されるコンフィグ構造体を生成する関数をセットする。
- SetNewCommandParamFunc : プラグインに渡されるコマンドパラメータ構造体を生成する関数をセットする。
- SetNewEventResultFunc  : プラグインを呼び出す側でイベントを処理した結果の構造体を生成する関数をセットする。

### イベント関数
- EventEmit : イベントを発生させる

### Plugin インターフェイス
PluginHandleと対応している。

- Init    : プラグインの初期化をする。
- Start   : プラグインを開始する。
- Stop    : プラグインを停止する。
- Reload  : プラグインを再設定をする。
- Fini    : プラグインの終了処理をする。
- Command : プラグインに送られたコマンドを処理する。
 
## 注意事項
- 1. plugin呼び出し側とplugin側で共有される。結果の構造体やコンフィグ構造体やコマンドパラメータ構造体は、バイト列に変換して渡しているため、encode/decodeができるように、必要なメンバの先頭は大文字にしておく必要があります。
- 2. pluggerのソースコードが変更された場合は、importしているプログラムの再ビルドが必要です

## 使用例
使用例は example 参照
