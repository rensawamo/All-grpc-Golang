## Procotcol Buffers を用いた開発の進め方
1. スキーマの定義
2. 開発言語のオブジェクトを自動生成 (protoファイル)
3. バイナリ形式へシリアライズ (pdファイル)


## gRPC を用いた開発の進め方
上記のコードから クライアントとサーバを実装 (クライアント、 サーバファイル)

## gRPC によるクライエントとサーバ通信

主な通信の方式
1. Unary RPC (1req 1 res)
2. サーバーストリーミング(1req 多res)
3. クライアントストリーミング(多req 1res)
4. 双方向ストリーミング (多req 多res)


##   その他技術

#### Interceptor
メソッドの前後に処理を行うための仕組み（リクエストを受け取る前、レスポンスを返した後のタイミングなどで任意の処理を割り込ませるなど）


#### ssl通信
認証局が発行した 証明書を本番の環境ではもちいないといけない
