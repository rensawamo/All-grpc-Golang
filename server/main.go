package main

//  RPC サーバの実装

import (
	"bytes"
	"context"
	"fmt"
	"grpc-lesson/pb"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

// pbのサーバをポインタに持つ サーバでの関数群を定義
type server struct {
	pb.UnimplementedFileServiceServer
}

// １つのリクエストに対して 1つのレスポンスを返す関数を定義
func (*server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	fmt.Println("ListFiles was invoked")

	dir := "/home/rensawamo/desktop/grpc-searver/storage"

	paths, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, path := range paths {
		if !path.IsDir() {
			filenames = append(filenames, path.Name())
		}
	}
	res := &pb.ListFilesResponse{
		Filenames: filenames,
	}
	return res, nil

}

// １つのリクエストに対して 複数の レスポンスをおくる関数の定義
func (*server) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {
	fmt.Println("Download was invoked")

	filename := req.GetFilename()
	path := "/home/rensawamo/desktop/grpc-searver/storage/" + filename

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 5)
	for {
		n, err := file.Read(buf)     // n は読み込んだバイト数が返される
		if n == 0 || err == io.EOF { // なにも読み込まれなかった or ファイルの最後まで到達した
			break
		}
		if err != nil {
			return err
		}
		// 読み込んだ内容を レスポンスに詰める
		res := &pb.DownloadResponse{Data: buf[:n]}
		sendErr := stream.Send(res)
		if sendErr != nil {
			return sendErr
		}

		time.Sleep(1 * time.Second)
	}
	return nil // resoposeが終了する
}

// 複数のリクエストをさばいて １つのレスポンスを送信する
func (*server) Upload(stream pb.FileService_UploadServer) error {
	fmt.Println("Upload was invoked")
	var buf bytes.Buffer
	for {
		req, err := stream.Recv() //クライアントから 複数のレスポンスを取得することが可能
		if err == io.EOF {        // eofの処理は ioに任せる
			res := &pb.UploadResponse{Size: int32(buf.Len())}
			return stream.SendAndClose(res)
		}
		if err != nil {
			return err
		}
		data := req.GetData()
		log.Printf("received data(bytes): %v", data)
		log.Printf("received data(string): %v", string(data))
		buf.Write(data)
	}
}

// 複数対複数
func (*server) UploadAndNotifyProgress(stream pb.FileService_UploadAndNotifyProgressServer) error {
	fmt.Println("UploadAndNotifyProgress was invoked")
	size := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		data := req.GetData()
		log.Printf("received data: %v", data)
		size += len(data)

		res := &pb.UploadAndNotifyProgressResponse{
			Msg: fmt.Sprintf("received %vbytes", size),
		}
		err = stream.Send(res)
		if err != nil {
			return err
		}
	}
}

// サーバ側のインターセプタを実装
func logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		log.Printf("request data: %+v", req)
		resp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}
		log.Printf("response data: %+v", resp)

		return resp, nil
	}
}

func main() {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// s := grpc.NewServer()
	s := grpc.NewServer(grpc.UnaryInterceptor(logging()))  // 今回は Unaryのログを挟んでみた
	pb.RegisterFileServiceServer(s, &server{}) // grpc サーバにファイル を登録する

	fmt.Println("server is running")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
