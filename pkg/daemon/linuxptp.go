package daemon

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
)

const (
	PTP4L_CONF_FILE_PATH = "/etc/ptp4l.conf"
)

type linuxPTPProcessManager struct {
	process	[]*ptpProcess
}

type ptpProcess struct {
	name	string
	exitCh	chan bool
	cmd	*exec.Cmd
}

// linuxPTPUpdate 
type linuxPTPUpdate struct {
	updateCh	chan bool
	nodeProfile	*ptpv1.NodePTPProfile
}

type linuxPTP struct {
	// node name where daemon is running
	nodeName	string
	namespace	string

	ptpUpdate	*linuxPTPUpdate
	// channel ensure LinuxPTP.Run() exit when main function exits.
	// stopCh is created by main function and passed by Daemon via NewLinuxPTP()
	stopCh <-chan struct{}
}

func NewLinuxPTP(
	nodeName	string,
	namespace	string,
	ptpUpdate	*linuxPTPUpdate,
	stopCh		<-chan struct{},
) *linuxPTP {
	return &linuxPTP{
		nodeName:	nodeName,
		namespace:	namespace,
		ptpUpdate:	ptpUpdate,
		stopCh:		stopCh,
	}
}

func (lp *linuxPTP) Run() {
	processManager := &linuxPTPProcessManager{}
	for {
		select {
		case <-lp.ptpUpdate.updateCh:
			err := applyNodePTPProfile(processManager, lp.ptpUpdate.nodeProfile)
			if err != nil {
				logging.Errorf("linuxPTP apply node profile failed: %v", err)
			}
		case <-lp.stopCh:
			logging.Debugf("linuxPTP stop signal received, existing..")
			return
		}
	}
	return
}

func applyNodePTPProfile(pm *linuxPTPProcessManager, nodeProfile *ptpv1.NodePTPProfile) error {
	logging.Debugf("applyNodePTPProfile() NodePTPProfile: %+v", nodeProfile)
	for _, p := range pm.process {
		if p != nil {
			cmdStop(p)
			p = nil
		}
	}

	pm.process = append(pm.process, &ptpProcess{
			name: "phc2sys",
			exitCh: make(chan bool),
			cmd: phc2sysCreateCmd(nodeProfile)})

	pm.process = append(pm.process, &ptpProcess{
			name: "ptp4l",
			exitCh: make(chan bool),
			cmd: ptp4lCreateCmd(nodeProfile)})

	for _, p := range pm.process {
		if p != nil {
			go cmdRun(p)
		}
	}
	return nil
}

func phc2sysCreateCmd(nodeProfile *ptpv1.NodePTPProfile) *exec.Cmd {
	cmdLine := fmt.Sprintf("/usr/sbin/phc2sys %s", nodeProfile.Phc2sysOpts)
	args := strings.Split(cmdLine, " ")
	return exec.Command(args[0], args[1:]...)
}

func ptp4lCreateCmd(nodeProfile *ptpv1.NodePTPProfile) *exec.Cmd {
	cmdLine := fmt.Sprintf("/usr/sbin/ptp4l -m -f %s -i %s %s",
			PTP4L_CONF_FILE_PATH,
			strings.Join(nodeProfile.Interfaces[:], " "),
			nodeProfile.Ptp4lOpts)
	args := strings.Split(cmdLine, " ")
	return exec.Command(args[0], args[1:]...)
}

func cmdRun(p *ptpProcess) {
	logging.Debugf("Starting %s...", p.name)
	logging.Debugf("%s cmd: %+v", p.name, p.cmd)

	defer func() {
		p.exitCh <- true
	}()

	cmdReader, err := p.cmd.StdoutPipe()
	if err != nil {
		logging.Errorf("cmdRun() error creating StdoutPipe for %s: %v", p.name, err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = p.cmd.Start()
	if err != nil {
		logging.Errorf("cmdRun() error starting %s: %v", p.name, err)
		return
	}

	err = p.cmd.Wait()
	if err != nil {
		logging.Errorf("cmdRun() error waiting for %s: %v", p.name, err)
		return
	}
	return
}

func cmdStop (p *ptpProcess) {
	logging.Debugf("Stopping %s...", p.name)
	if p.cmd == nil {
		return
	}

	if p.cmd.Process != nil {
		logging.Debugf("Sending TERM to PID: %d", p.cmd.Process.Pid)
		p.cmd.Process.Signal(syscall.SIGTERM)
	}

	<-p.exitCh
	logging.Debugf("Process %d terminated", p.cmd.Process.Pid)
}
