package main

import (
	"context"
	"log"
	"net"
	"regexp"

	"bufio"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/grpc"

	pb "ifundrobot/protos"

	"google.golang.org/grpc/reflection"
)

const (
	port                 = ":50051"
	supervisorConfigPath = "/etc/supervisor/conf.d/"
	configTemplate       = `[program:%name%_%currency%]
command=/home/john/SuperFundingBot/.venv/bin/python /home/john/bitfinex-funding-robot/create_funding_offers3.py %name% -s %currency%
autostart=true
autorestart=true
stderr_logfile=/var/log/ifund/%name%_%currency%.err.log
stdout_logfile=/var/log/ifund/%name%_%currency%.out.log
user=john
`
	supervisorController = "/usr/bin/supervisorctl"
)

type server struct {
	pb.UnimplementedIfundrobotServer
}

func (s *server) CreateRobot(ctx context.Context, in *pb.RobotRequest) (*pb.CreateReply, error) {
	log.Printf("CreateRobot User: %s, Currency: %s", in.Name, in.Currency)

	if !createSupervisorConfig(in.Name, in.Currency) {
		return &pb.CreateReply{Code: 1, Message: "create failed"}, nil
	}

	if !updateSupervisor() {
		return &pb.CreateReply{Code: 1, Message: "update failed"}, nil
	}

	return &pb.CreateReply{Message: "create success"}, nil
}

func (s *server) RobotStatus(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {
	supervisorStates := []string{"STARTING", "STOPPED", "RUNNING"}

	state := robotState(in.Name, in.Currency)
	code := 0

	log.Printf("Robot state %s", state)
	if !contains(supervisorStates, state) {
		code = 1
	}
	return &pb.StatusReply{Code: int32(code), State: state}, nil
}

func (s *server) StopRobot(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {

	result, err := robotDoAction(in.Name, in.Currency, "stop")
	if err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: ""}, err
	}

	log.Println(result)
	if strings.Contains(result, "ERROR") {
		return &pb.StatusReply{Code: 1, State: "", Message: result}, err
	}
	return &pb.StatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *server) StartRobot(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {
	result, err := robotDoAction(in.Name, in.Currency, "stop")
	if err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: ""}, err
	}
	log.Println(result)
	if strings.Contains(result, "ERROR") {
		return &pb.StatusReply{Code: 1, State: "", Message: result}, err
	}
	return &pb.StatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *server) RestartRobot(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {
	result, err := robotDoAction(in.Name, in.Currency, "restart")
	if err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: result}, err
	}

	log.Println(result)
	return &pb.StatusReply{Code: 0, State: "", Message: result}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterIfundrobotServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Println("Server is running...")
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func createSupervisorConfig(username, currency string) bool {
	configFilename := robotFilename(username, currency)
	if fileExists(configFilename) {
		return true
	}

	tmpStr1 := strings.ReplaceAll(configTemplate, "%name%", username)
	configContent := strings.ReplaceAll(tmpStr1, "%currency%", currency)

	log.Println(configContent)

	f, err := os.Create(configFilename)
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)
	n, err := w.WriteString(configContent)
	check(err)
	w.Flush()
	log.Printf("wrote %d bytes\n", n)
	return true
}

func updateSupervisor() bool {
	out, err := exec.Command(supervisorController, "update").Output()
	// if there is an error with our execution
	// handle it here
	if err != nil {
		log.Printf("%s", err)
	}
	log.Println(out)
	return true
}

func robotDoAction(username, currency, action string) (string, error) {
	serviceName := robotServiceName(username, currency)
	out, err := exec.Command(supervisorController, action, serviceName).Output()
	if err != nil {
		log.Printf("%s", err)
		return "", err
	}
	log.Printf("Robot action result: %s\n", out)
	return string(out), nil
}

func robotState(username, currency string) string {
	serviceName := robotServiceName(username, currency)

	log.Printf("Supervisor service name: %s\n", serviceName)

	out, err := exec.Command(supervisorController, "status", serviceName).Output()
	if err != nil {
		log.Printf("%s", err)
		return ""
	}
	log.Printf("Status: %s\n", out)
	pattern := regexp.MustCompile(`(?P<name>[a-zA-Z]+).*(?P<status>(STARTING|STOPPED|RUNNING)+).*`)
	if !pattern.MatchString(string(out)) {
		return ""
	}

	matches := pattern.FindStringSubmatch(string(out))
	lastIndex := pattern.SubexpIndex("status")
	log.Printf("last => %d\n", lastIndex)
	return matches[lastIndex]
}

func robotFilename(username, currency string) string {
	return supervisorConfigPath + username + "_" + currency + ".conf"
}

func robotServiceName(username, currency string) string {
	return username + "_" + currency
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
