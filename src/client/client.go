package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
	"log"
	"math/rand"

	pb "github.com/matiasinsaurralde/go-grpc-bidirectional-streaming-example/src/proto"

	"time"

	"google.golang.org/grpc"
)

const (
	defaultTarget     = "localhost:50005"
	defaultCertOption = "cert/ca-cert.pem"
)

var (
	targetOption   = ""
	insecureOption bool
	certOption     string
	tlsMode        bool
)

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile(certOption)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	if insecureOption {
		config.InsecureSkipVerify = true
	}

	return credentials.NewTLS(config), nil
}

func main() {
	rand.Seed(time.Now().Unix())

	flag.StringVar(&targetOption, "server", defaultTarget, "set the gRPC target server")
	flag.BoolVar(&insecureOption, "insecure", false, "use insecure mode")
	flag.StringVar(&certOption, "cert", defaultCertOption, "certificate path")
	flag.BoolVar(&tlsMode, "tls", false, "use TLS")

	flag.Parse()
	fmt.Printf("Target server is: %s\n", targetOption)
	fmt.Printf("Insecure mode: %v\n", insecureOption)
	fmt.Printf("TLS mode: %v\n", tlsMode)
	if tlsMode {
		fmt.Printf("Certificate path: %s\n", certOption)
	}

	var conn *grpc.ClientConn
	var err error
	if tlsMode {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}
		conn, err = grpc.Dial(targetOption, grpc.WithTransportCredentials(tlsCredentials))
		if err != nil {
			log.Fatal("cannot dial server: ", err)
		}
	} else {
		conn, err = grpc.Dial(targetOption, grpc.WithInsecure())
		if err != nil {
			log.Fatal("cannot dial server: ", err)
		}
	}

	// create stream
	client := pb.NewMathClient(conn)
	stream, err := client.Max(context.Background())
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}

	var max int32
	ctx := stream.Context()
	done := make(chan bool)

	// first goroutine sends random increasing numbers to stream
	// and closes int after 10 iterations
	go func() {
		for i := 1; i <= 10; i++ {
			// generate random nummber and send it to stream
			rnd := int32(rand.Intn(i))
			req := pb.Request{Num: rnd}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("can not send %v", err)
			}
			log.Printf("%d sent", req.Num)
			time.Sleep(time.Millisecond * 200)
		}
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	// second goroutine receives data from stream
	// and saves result in max variable
	//
	// if stream is finished it closes done channel
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			max = resp.Result
			log.Printf("new max %d received", max)
		}
	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
	log.Printf("finished with max=%d", max)
}
