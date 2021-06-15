package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	//xContext "golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

func main() {
	//ctx, cancel := xContext.WithCancel(xContext.Background())
	ctx, cancel := context.WithCancel(context.Background())
	group, errCtx := errgroup.WithContext(ctx)

	srv := &http.Server{Addr: ":8080"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world\n")
	})

	// 开启 http server 线程
	group.Go(func() error {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
			return err
		}
		return errors.New("srv.ListenAndServe() over")
	})
	// 监听errCtx
	group.Go(func() error {
		select {
		case <-errCtx.Done():
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("Httpserver: Shutdown() error: %s", err)
				return err
			}
			log.Printf("Httpserver: Shutdown() succeed")
		}
		return nil
	})

	// 监听linux signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	//group.Go(func() error {
	//	for {
	//		s := <-c
	//		log.Printf("get a signal %s", s.String())
	//		switch s {
	//		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
	//			cancel()
	//			log.Printf("cancel exec")
	//			//time.Sleep(time.Second)
	//			return nil
	//		case syscall.SIGHUP:
	//		default:
	//			return nil
	//		}
	//	}
	//})

	group.Go(func() error {
		for {
			select {
			case s := <-c:
				log.Printf("get a signal %s ", s.String())
				//这样别的协程就可以通过errCtx获取到err信息，以便决定是否需要取消后续操作
				cancel()
				log.Printf("cancel exec")
				return nil
			default:

			}
		}
	})

	// 捕获err
	err := group.Wait()
	if err == nil {
		fmt.Println("都完成了")
	} else {
		fmt.Printf("get error:%v\n", err)
	}
}
