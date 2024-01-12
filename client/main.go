package main  //client も同じく  main パッケージ

import (
	"context"
	"fmt"
	"grpc-lesson/pb"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
)

// クライアントでは、 pbによって定義された 通信方式のインターフェイスの関数が実装される

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn) // ファイルサービスクライアントの取得
	// callListFiles(client)
	// callDownload(client)
	// CallUpload(client)
	CallUploadAndNotifyProgress(client)
}

// 1 リクエスト 1 レスポンスの gRPCの通信の方式
func callListFiles(client pb.FileServiceClient) {
	res, err := client.ListFiles(context.Background(), &pb.ListFilesRequest{})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(res.GetFilenames())
}


// １リクエスト 複数のレスポンス方式
func callDownload(client pb.FileServiceClient) {
	req := &pb.DownloadRequest{Filename: "name.txt"}
	stream, err := client.Download(context.Background(), req)  // インターフェイスで stream が実装されている
	if err != nil {
		log.Fatalln(err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Response from Download(bytes): %v", res.GetData())
		log.Printf("Response from Download(string): %v", string(res.GetData()))
	}
}

// 複数の リクエスト と１レスポンス方式
func CallUpload(client pb.FileServiceClient) {
	filename := "sports.txt"
	path := "/home/rensawamo/desktop/grpc-searver/storage" + filename

	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	buf := make([]byte, 5)
	for {
		n, err := file.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		req := &pb.UploadRequest{Data: buf[:n]}
		sendErr := stream.Send(req)
		if sendErr != nil {
			log.Fatalln(sendErr)
		}
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("received data size: %v", res.GetSize())
}

// 双方向ストリーミングのリクエスト処理
func CallUploadAndNotifyProgress(client pb.FileServiceClient) {
	filename := "sports.txt"
	path := "/home/rensawamo/desktop/grpc-searver/storage/" + filename

	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	stream, err := client.UploadAndNotifyProgress(context.Background()) // streamの取得
	if err != nil {
		log.Fatalln(err)
	}

	// 双方向ストリーミングでは、並行でリクエストとレスポンスをうけとっていく
	// go ルーチンを使用

	// request
	buf := make([]byte, 5)
	go func() {
		for {
			n, err := file.Read(buf)
			if n == 0 || err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalln(err)
			}

			req := &pb.UploadAndNotifyProgressRequest{Data: buf[:n]}
			sendErr := stream.Send(req)
			if sendErr != nil {
				log.Fatalln(sendErr)
			}
			time.Sleep(1 * time.Second)
		}

		err := stream.CloseSend() 
		if err != nil {
			log.Fatalln(err)
		}
	}()
	// response
	ch := make(chan struct{})
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF { // requestの stream.CloseSend()の実行により通知を受け取る
				break
			}
			if err != nil {
				log.Fatalln(err)
			}

			log.Printf("received message: %v", res.GetMsg())
		}
		close(ch)
	}()

	<-ch //ここまで到達すると  close(ch) に到達するので go ルーチンで処理がまわっている間まで の双方向ストリーム保障される

}
