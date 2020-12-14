package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"net"

	pb "github.com/matiasinsaurralde/go-grpc-bidirectional-streaming-example/src/proto"

	"google.golang.org/grpc"
)

const (
	defaultListenAddr = "127.0.0.1:50005"
	defaultCertPath   = "cert/server-cert.pem"
	defaultKeyPath    = "cert/server-key.pem"
)

var (
	certPath string
	keyPath  string
	tlsMode  bool
)

type server struct{}

func (s server) Max(srv pb.Math_MaxServer) error {
	log.Println("start new server")
	var max int32
	ctx := srv.Context()

	for {

		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream
		req, err := srv.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}

		// continue if number reveived from stream
		// less than max
		if req.Num <= max {
			continue
		}

		// update max and send it to stream
		max = req.Num
		resp := pb.Response{Result: max}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
		log.Printf("send new max=%d", max)
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	listenAddrPtr := flag.String("listen", defaultListenAddr, "set the gRPC listen address")
	flag.BoolVar(&tlsMode, "tls", false, "Use TLS mode")
	flag.StringVar(&certPath, "cert", defaultCertPath, "Server certificate path")
	flag.StringVar(&keyPath, "key", defaultKeyPath, "Server key path")
	flag.Parse()

	fmt.Printf("Listen address is: %s\n", *listenAddrPtr)
	fmt.Printf("TLS Mode: %v\n", tlsMode)
	if tlsMode {
		fmt.Printf("Server certificate path: %s\n", certPath)
		fmt.Printf("Server key path: %s\n", keyPath)
	}

	// create listener
	lis, err := net.Listen("tcp", *listenAddrPtr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var s *grpc.Server
	if tlsMode {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}
		s = grpc.NewServer(grpc.Creds(tlsCredentials))
	} else {
		s = grpc.NewServer()
	}

	pb.RegisterMathServer(s, server{})

	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
