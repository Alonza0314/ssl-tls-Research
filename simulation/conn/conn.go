package conn

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"simulation/config"
	"simulation/pki"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type ListenerInterface func(string, HandlerInterface)

type HandlerInterface func(conn net.Conn, ctx context.Context)

func TCPListener(serverCfg string, handler HandlerInterface) {
	log.Println("=> Server Starting...")

	viper.SetConfigFile(serverCfg)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config:\n\t%s\n", err.Error())
	}

	host, port := viper.GetString("server.addr"), viper.GetInt("server.port")

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%v", host, port))
	if err != nil {
		log.Fatalf("failed to listen on %s:%v:\n\t%s\n", host, port, err.Error())
	}
	defer listener.Close()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println()
		log.Println("==> Server Stopping...")
		cancel()
		listener.Close()
	}()

	log.Println("===== Server Started =====")
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			log.Println("===== Server Stopped =====")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if ne, ok := err.(*net.OpError); ok && ne.Err.Error() == "use of closed network connection" {
					continue
				}
				log.Printf("failed to accept connection:\n\t%s\n", err.Error())
				continue
			}
			log.Println(config.NEW, "CONNECTION from", conn.RemoteAddr())
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				defer conn.Close()
				handler(conn, ctx)
			}(conn)
		}
	}
}

func TCPHandler(conn net.Conn, ctx context.Context) {
	if err := conn.SetReadDeadline(time.Now().Add(config.CONN_TIMEOUT)); err != nil {
		log.Printf("failed to set read timeoute:\n\t%s\n", err.Error())
		return
	}
	var req request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		log.Printf("failed to decode request:\n\t%s\n", err.Error())
		return
	}
	log.Println(config.RECEIVED, config.CLIENT_HELLO, "from", conn.RemoteAddr())
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		log.Printf("failed to reset read timeout:\n\t%s\n", err.Error())
		return
	}

	viper.SetConfigFile("../config/servercfg.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("failed to read config:\n\t%s\n", err.Error())
		return
	}
	// keyPair, err := pki.NewKeyPair(viper.GetString("server.privateKey"), viper.GetString("server.publicKey"))
	// if err != nil {
	// 	log.Printf("failed to new keypair:\n\t%s\n", err.Error())
	//	return
	// }

	// rx, tx, err := keyPair.SessionKeys(req.PublicKey)
	// if err != nil {
	// 	log.Printf("failed to compute rx tx:\n\t%s\n", err.Error())
	// 	return
	// }

	keyPair, err := pki.NewKeyPair()
	if err != nil {
		log.Printf("failed to new keypair:\n\t%s\n", err.Error())
		return
	}

	rx, tx, err := keyPair.ServerSessionKeys(req.PublicKey)
	if err != nil {
		log.Printf("failed to compute rx tx:\n\t%s\n", err.Error())
		return
	}

	txHeader := make([]byte, config.STREAM_HEADER_SIZE)
	if _, err := rand.Read(txHeader); err != nil {
		log.Printf("failed to make txHeader:\n\t%s\n", err.Error())
		return
	}

	sender, err := pki.NewEncryptor(tx, txHeader)
	if err != nil {
		log.Printf("failed to new encryptor:\n\t%s\n", err.Error())
		return
	}

	receiver, err := pki.NewDecryptor(rx, []byte(req.Options[config.TX_HEADER]))
	if err != nil {
		log.Printf("failed to new decryptor:\n\t%s\n", err.Error())
		return
	}
	options := map[string]string{
		config.TX_HEADER: string(txHeader),
	}
	rep, err := NewReply(config.SERVER_HELLO, keyPair.Public(), options)
	if err != nil {
		log.Printf("failed to new reply:\n\t%s\n", err.Error())
		return
	}

	if err = rep.SendReply(conn); err != nil {
		log.Printf("failed to send reply:\n\t%s\n", err.Error())
		return
	}

	// TODO
	// ctx graceful stop
	for {
		select {
		case <-ctx.Done():
		default:
			// TODO
			// Read and Write
			// if Read EOF, break for and return
		}
	}
	fmt.Println(sender, receiver)
}
