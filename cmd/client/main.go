package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	pb "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

const (
	// TODO: Make configurable
	timeout = 10 * time.Second
)

func getClientConnection(context *cli.Context) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(context.GlobalString("connect"), grpc.WithInsecure(), grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	return conn, nil
}

func openFile(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config at %s not found", path)
		}
		return nil, err
	}
	return f, nil
}

func loadPodSandboxConfig(path string) (*pb.PodSandboxConfig, error) {
	f, err := openFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config pb.PodSandboxConfig
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func loadContainerConfig(path string) (*pb.ContainerConfig, error) {
	f, err := openFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config pb.ContainerConfig
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "ocic"
	app.Usage = "client for ocid"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		podSandboxCommand,
		containerCommand,
		runtimeVersionCommand,
		imageCommand,
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "connect",
			Value: "/var/run/ocid.sock",
			Usage: "Socket to connect to",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
