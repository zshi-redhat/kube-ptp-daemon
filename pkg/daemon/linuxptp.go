package daemon

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/zshi-redhat/kube-ptp-daemon/logging"
	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
)

const (
	PTP4L_CONF_FILE_PATH = "/etc/ptp4l.conf"
)

// linuxPTPProcessManager manages a set of ptpProcess
// which could be ptp4l, phc2sys or timemaster.
// Processes in linuxPTPProcessManager will be started
// or stopped simultaneously.
type linuxPTPProcessManager struct {
	process	[]*ptpProcess
}

type ptpProcess struct {
	name	string
	exitCh	chan bool
	cmd	*exec.Cmd
}

// linuxPTPUpdate controls whether to update linuxPTP conf
// and contains linuxPTP conf to be updated. It's rendered
// and passed to linuxptp instance by daemon.
type linuxPTPUpdate struct {
	updateCh	chan bool
	nodeProfile	*ptpv1.NodePTPProfile
}

// linuxPTP is the main structure for linuxptp instance.
// It contains all the necessary data to run linuxptp instance.
type linuxPTP struct {
	// node name where daemon is running
	nodeName	string
	namespace	string

	ptpUpdate	*linuxPTPUpdate
	// channel ensure LinuxPTP.Run() exit when main function exits.
	// stopCh is created by main function and passed by Daemon via NewLinuxPTP()
	stopCh <-chan struct{}
}

// NewLinuxPTP is called by daemon to generate new linuxptp instance
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

// Run in a for loop to listen for any linuxPTPUpdate changes
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
			for _, p := range processManager.process {
				if p != nil {
					cmdStop(p)
					p = nil
				}
			}
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
			logging.Debugf("stopping commands.... %+v", p)
			cmdStop(p)
			p = nil
		}
	}

	// All process should have been stopped,
	// clear process in process manager.
	// Assigning pm.process to nil releases
	// the underlying slice to the garbage
	// collector (assuming there are no other
	// references).
	pm.process = nil

	// TODO:
	// compare nodeProfile with previous config,
	// only apply when nodeProfile changes

	if nodeProfile.Phc2sysOpts != nil {
		pm.process = append(pm.process, &ptpProcess{
			name: "phc2sys",
			exitCh: make(chan bool),
			cmd: phc2sysCreateCmd(nodeProfile)})
		logging.Debugf("applyNodePTPProfile() not starting phc2sys, phc2sysOpts empty")
	}

	if nodeProfile.Ptp4lOpts != nil && nodeProfile.Interface != nil {
		pm.process = append(pm.process, &ptpProcess{
			name: "ptp4l",
			exitCh: make(chan bool),
			cmd: ptp4lCreateCmd(nodeProfile)})
		logging.Debugf("applyNodePTPProfile() not starting ptp4l, ptp4lOpts or interface empty")
	}

	for _, p := range pm.process {
		if p != nil {
			time.Sleep(1*time.Second)
			go cmdRun(p)
		}
	}
	return nil
}

// phc2sysCreateCmd generate phc2sys command
func phc2sysCreateCmd(nodeProfile *ptpv1.NodePTPProfile) *exec.Cmd {
	cmdLine := fmt.Sprintf("/usr/sbin/phc2sys %s", *nodeProfile.Phc2sysOpts)
	args := strings.Split(cmdLine, " ")
	return exec.Command(args[0], args[1:]...)
}

// ptp4lCreateCmd generate ptp4l command
func ptp4lCreateCmd(nodeProfile *ptpv1.NodePTPProfile) *exec.Cmd {
	cmdLine := fmt.Sprintf("/usr/sbin/ptp4l -m -f %s -i %s %s",
		PTP4L_CONF_FILE_PATH,
		*nodeProfile.Interface,
		*nodeProfile.Ptp4lOpts)

	args := strings.Split(cmdLine, " ")
	return exec.Command(args[0], args[1:]...)
}


// cmdRun runs given ptpProcess and wait for errors
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

	done := make(chan struct{})

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
		done <- struct{}{}
	}()

	err = p.cmd.Start()
	if err != nil {
		logging.Errorf("cmdRun() error starting %s: %v", p.name, err)
		return
	}

	<-done

	err = p.cmd.Wait()
	if err != nil {
		logging.Errorf("cmdRun() error waiting for %s: %v", p.name, err)
		return
	}
	return
}

// cmdStop stops ptpProcess launched by cmdRun
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
